package usecase

import (
	"context"
	"driver/domain"
	"driver/interface/repository"
	"fmt"
	"log"
	"time"

	"github.com/tonkeeper/tongo"
	"github.com/tonkeeper/tongo/boc"
	"github.com/tonkeeper/tongo/liteapi"
	tgwallet "github.com/tonkeeper/tongo/wallet"
)

const (
	OpcodeSaveCoin  = uint32(0x7f30ee55)
	OpcodeStakeCoin = uint32(0x00)
)

type JettonWalletInteractor struct {
	client            *liteapi.Client
	jwalletRepository *repository.JettonWalletRepository
	driverWallet      *tgwallet.Wallet
}

func NewJettonWalletInteractor(client *liteapi.Client,
	jwalletRepository *repository.JettonWalletRepository,
	driverWallet *tgwallet.Wallet) *JettonWalletInteractor {
	interactor := &JettonWalletInteractor{
		client:            client,
		jwalletRepository: jwalletRepository,
		driverWallet:      driverWallet,
	}
	return interactor
}

func (interactor *JettonWalletInteractor) ExtractJettonWallets(accountId tongo.AccountID) (map[string]domain.JettonWallet, error) {

	var FoundWallets = make(map[string]domain.JettonWallet, 50)

	// Find las transactions of the account
	trans, err := interactor.client.GetLastTransactions(context.Background(), accountId, 50)
	if err != nil {
		fmt.Printf("Failed to get last transactions - %v\n", err.Error())
		return nil, err
	}

	var lastHash tongo.Bits256
	for err == nil && len(trans) > 0 {
		wallets := findDestByOpcode(trans, OpcodeSaveCoin)
		for _, w := range wallets {
			FoundWallets[w.Address] = w
		}

		// Extract the Lt and the Hash of last transaction
		lastLt := trans[len(trans)-1].Lt
		lastHash.FromHex(trans[len(trans)-1].Hash().Hex())

		// Extract previous transactions. GetTransactions function returns 16 items by max, this is why 16 is passed for the count parameter.
		trans, err = interactor.client.GetTransactions(context.Background(), 16, accountId, lastLt, lastHash)
		if err != nil {
			log.Printf("Failed to get transactions - %v\n", err.Error())
			fmt.Printf("❌ No wallet is kept due to error: %v", err.Error())
			return nil, err
		}

		// Remove the first element as it's already processed in previous loop
		if len(trans) > 0 {
			trans = trans[1:]
		}
	}

	return FoundWallets, nil
}

func (interactor *JettonWalletInteractor) Store(wallets map[string]domain.JettonWallet) error {
	for _, wallet := range wallets {
		_, err := interactor.jwalletRepository.InsertIfNotExists(wallet.Address, wallet.Info)
		if err != nil {
			log.Printf("Failed to insert jetton wallet record - %v\n", err.Error())
			return err
		}
	}
	return nil
}

func (interactor *JettonWalletInteractor) LoadNotNotified() ([]domain.JettonWallet, error) {

	wallets, err := interactor.jwalletRepository.FindAllNotNotified()
	if err != nil {
		log.Printf("Failed to load jetton wallet records - %v\n", err.Error())
		return nil, err
	}

	return wallets, nil
}

func (interactor *JettonWalletInteractor) SendMessageToJettonWallets(wallets []domain.JettonWallet) {
	// @TODO: The round-since value must be considered as a condition whether to call stakeCoin or not.
	//
	//	start from sooner roundSince
	//  run get_method get_treasury_state
	//  extract participations field
	//  check if the desired round-since does exist in participations or not
	//  if exists, skip it, otherwise send stake-coin message
	for _, wallet := range wallets {
		accid, err := tongo.AccountIDFromBase64Url(wallet.Address)
		if err != nil {
			log.Printf("Failed to parse wallet address %v - %v\n", wallet.Address, err.Error())
			continue
		}

		err = interactor.stakeCoin(accid)
		if err != nil {
			log.Printf("Failed to stake coin for wallet address %v - %v\n", wallet.Address, err.Error())
			continue
		} else {
			log.Printf("Successfully stake coin for wallet address %v.\n", wallet.Address)
			interactor.jwalletRepository.UpdateNotified(wallet.Address, time.Now())
		}
	}
}

func (interactor *JettonWalletInteractor) stakeCoin(accid tongo.AccountID) error {
	opcode := uint64(0x4cae3ab1)

	queryId := uint64(time.Now().Unix())
	roundSince := uint64(0) // @TODO: round-since value must be evaluated here

	cell := boc.NewCell()
	cell.WriteUint(opcode, 32)     // opcode
	cell.WriteUint(queryId, 64)    // query id
	cell.WriteUint(roundSince, 32) // round since
	cell.WriteUint(0, 2)           // return excess

	msg := tgwallet.Message{
		Amount:  100000000, //  tlb.Grams
		Address: accid,     //  tongo.AccountID
		Body:    cell,      //  *boc.Cell
		Code:    nil,       //  *boc.Cell
		Data:    nil,       //  *boc.Cell
		Bounce:  true,      //  bool
		Mode:    0,         //  uint8
	}

	err := interactor.driverWallet.Send(context.Background(), msg)
	if err != nil {
		log.Printf("Failed to send message - %v\n", err.Error())
		return err
	}

	return nil
}

func findDestByOpcode(trans []tongo.Transaction, opcode uint32) []domain.JettonWallet {
	wallets := make([]domain.JettonWallet, 0, 1)
	for _, t := range trans {
		ht := domain.NewHTransaction(&t.Transaction)
		dests := ht.GetDestByOpcode(opcode)
		info := domain.RelatedTransactionInfo{
			Value: ht.Value(),
			Time:  ht.UnixTime(),
			Hash:  ht.Hash().Base64(),
		}
		for _, accid := range dests {
			addr := accid.ToHuman(true, domain.IsTestNet())
			wallets = append(wallets, domain.JettonWallet{Address: addr, Info: info, CreateTime: time.Now()})
		}
	}

	return wallets
}
