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
	"github.com/tonkeeper/tongo/tlb"
	tgwallet "github.com/tonkeeper/tongo/wallet"
)

const (
	OpcodeSaveCoin  = uint32(0x7f30ee55)
	OpcodeStakeCoin = uint32(0x4cae3ab1)
)

type JettonWalletInteractor struct {
	client             *liteapi.Client
	memoInteractor     *MemoInteractor
	contractInteractor *ContractInteractor
	jwalletRepository  *repository.JettonWalletRepository
	driverWallet       *tgwallet.Wallet
}

func NewJettonWalletInteractor(client *liteapi.Client,
	memoInteractor *MemoInteractor,
	contractInteractor *ContractInteractor,
	jwalletRepository *repository.JettonWalletRepository,
	driverWallet *tgwallet.Wallet) *JettonWalletInteractor {
	interactor := &JettonWalletInteractor{
		client:             client,
		memoInteractor:     memoInteractor,
		contractInteractor: contractInteractor,
		jwalletRepository:  jwalletRepository,
		driverWallet:       driverWallet,
	}
	return interactor
}

func (interactor *JettonWalletInteractor) ExtractJettonWallets(treasuryAccount tongo.AccountID) (map[string]domain.JettonWallet, error) {

	var FoundWallets = make(map[string]domain.JettonWallet, 50)

	// Read the latest processed transaction info
	latestProcessedHash, err := interactor.memoInteractor.GetLatestProcessedHash()
	if err != nil {
		fmt.Printf("Failed to get last processed hash - %v\n", err.Error())
		return nil, err
	}

	// Get last transactions of the treasury account, sorted decently by time.
	trans, err := interactor.client.GetLastTransactions(context.Background(), treasuryAccount, 50)
	if err != nil {
		fmt.Printf("Failed to get last transactions - %v\n", err.Error())
		return nil, err
	}

	// Keep the first processing transaction's hash, so that in next call, stop searching through the processed transactions.
	var firstTrans *domain.HTransaction
	firstTransHash := ""
	if len(trans) > 0 {
		firstTrans = domain.NewHTransaction(&trans[0].Transaction)
		htf := domain.NewHTransactionFormatter(firstTrans)
		firstTransHash = htf.Hash()
	}

	// Start processing transactions
	var hash tongo.Bits256
	reachEnd := firstTransHash == latestProcessedHash
	if reachEnd {
		log.Printf("No new transaction to be processed.\n")
	}

	for err == nil && len(trans) > 0 && !reachEnd {
		index := findLastUnprocessed(trans, latestProcessedHash)
		reachEnd = index < len(trans)
		if reachEnd {
			trans = trans[0:index]
		}

		wallets := findDestByOpcode(trans, OpcodeSaveCoin)
		for _, w := range wallets {
			FoundWallets[w.Address] = w
		}

		// If the latest processed transaction is not reached,
		if !reachEnd {
			// Extract the Lt and the Hash of last transaction
			lt := trans[len(trans)-1].Lt
			hash.FromHex(trans[len(trans)-1].Hash().Hex())

			// Extract previous transactions. GetTransactions function returns 16 items by max, this is why 16 is passed for the count parameter.
			trans, err = interactor.client.GetTransactions(context.Background(), 16, treasuryAccount, lt, hash)
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
	}

	// Keep the first hash as the latest processed hash.
	if firstTransHash != "" && firstTransHash != latestProcessedHash {
		err = interactor.memoInteractor.SetLatestProcessedHash(firstTransHash)
		if err != nil {
			log.Printf("Failed to update latest processed hash - %v\n", err.Error())
		}
	}

	return FoundWallets, nil
}

func (interactor *JettonWalletInteractor) Store(wallets map[string]domain.JettonWallet) error {
	for _, wallet := range wallets {
		_, err := interactor.jwalletRepository.InsertIfNotExists(wallet.Address, wallet.RoundSince, wallet.Info)
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

func (interactor *JettonWalletInteractor) SendMessageToJettonWallets(wallets []domain.JettonWallet) error {
	// @TODO: The round-since value must be considered as a condition whether to call stakeCoin or not.
	//
	//	start from sooner roundSince
	//  run get_method get_treasury_state
	//  extract participations field
	//  check if the desired round-since does exist in participations or not
	//  if exists, skip it, otherwise send stake-coin message

	// Split the wallets based on their round-since
	splitted := make(map[uint32][]domain.JettonWallet, 0)
	for _, wallet := range wallets {
		roundSince := wallet.RoundSince
		if _, exist := splitted[roundSince]; exist {
			splitted[roundSince] = append(splitted[roundSince], wallet)
		} else {
			subList := make([]domain.JettonWallet, 0, 1)
			subList = append(subList, wallet)
			splitted[roundSince] = subList
		}
	}

	treasuryState, err := interactor.contractInteractor.GetTreasuryState()
	if err != nil {
		log.Printf("Failed to send message to wallets - %v\n", err.Error())
		return err
	}

	for roundSince, subList := range splitted {
		if _, exist := treasuryState.Participations[roundSince]; exist {
			continue
		}

		for _, wallet := range subList {
			accid, err := tongo.AccountIDFromBase64Url(wallet.Address)
			if err != nil {
				log.Printf("Failed to parse wallet address %v - %v\n", wallet.Address, err.Error())
				continue
			}

			// @TODO: add state for jetton-wallet record: new, ongoing, done, error
			err = interactor.stakeCoin(accid, roundSince)
			if err != nil {
				log.Printf("Failed to stake coin for wallet address %v - %v\n", wallet.Address, err.Error())
				continue
			} else {
				// @TODO: update the wallet regarding to round-since
				err = interactor.jwalletRepository.UpdateNotified(wallet.Address, roundSince, time.Now())
				if err != nil {
					log.Printf("Failed to update wallet address %v - %v\n", wallet.Address, err.Error())
					continue
				} else {
					// @TODO: organize log messages, and shorten them.
					log.Printf("Successfully stake coin for wallet address %v.\n", wallet.Address)
				}
			}
		}
	}

	return nil
}

func (interactor *JettonWalletInteractor) stakeCoin(accid tongo.AccountID, roundSince uint32) error {
	queryId := uint64(time.Now().Unix())

	cell := boc.NewCell()
	cell.WriteUint(uint64(OpcodeStakeCoin), 32) // opcode
	cell.WriteUint(queryId, 64)                 // query id
	cell.WriteUint(uint64(roundSince), 32)      // round since
	cell.WriteUint(0, 2)                        // return excess

	msg := tgwallet.Message{
		Amount:  100000000, //  tlb.Grams
		Address: accid,     //  tongo.AccountID
		Body:    cell,      //  *boc.Cell
		Code:    nil,       //  *boc.Cell
		Data:    nil,       //  *boc.Cell
		Bounce:  true,      //  bool
		Mode:    1,         //  uint8	/ Pay transfer fees separately from the message value /
	}

	err := interactor.driverWallet.Send(context.Background(), msg)
	if err != nil {
		log.Printf("Failed to send message - %v\n", err.Error())
		return err
	}

	return nil
}

func findLastUnprocessed(trans []tongo.Transaction, lastHash string) int {
	for i, t := range trans {
		ht := domain.NewHTransaction(&t.Transaction)
		htf := domain.NewHTransactionFormatter(ht)
		if htf.Hash() == lastHash {
			return i
		}
	}

	return len(trans)
}

func findDestByOpcode(trans []tongo.Transaction, opcode uint32) []domain.JettonWallet {
	wallets := make([]domain.JettonWallet, 0, 1)
	for _, t := range trans {
		ht := domain.NewHTransaction(&t.Transaction)

		// leave transaction if it's failed
		if !ht.IsSucceeded() {
			continue
		}

		msgs := ht.GetMessagesByOpcode(opcode)
		info := domain.RelatedTransactionInfo{
			Value: ht.Value(),
			Time:  ht.UnixTime(),
			Hash:  ht.Hash().Base64(),
		}
		for _, msg := range msgs {
			accid := msg.Dest()
			cell := msg.GetBody()
			m := domain.SaveCoinMessage{}
			tlb.Unmarshal(cell, &m)

			// log.Printf(
			// 	"\n"+
			// 		"        opcode = %x\n"+
			// 		"      query id = %x\n"+
			// 		"        amount = %v\n"+
			// 		"   round since =%v\n"+
			// 		" return excess = %v\n", m.Opcode, m.QuieryId, m.Amount, m.RoundSince, m.ReturnExcess)

			addr := accid.ToHuman(true, domain.IsTestNet())
			wallets = append(wallets, domain.JettonWallet{Address: addr, RoundSince: m.RoundSince, Info: info, CreateTime: time.Now()})
		}
	}

	return wallets
}
