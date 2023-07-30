package usecase

import (
	"context"
	"driver/domain"
	"driver/interface/repository"
	"log"
	"time"

	"github.com/tonkeeper/tongo"
	"github.com/tonkeeper/tongo/boc"
	"github.com/tonkeeper/tongo/liteapi"
	"github.com/tonkeeper/tongo/tlb"
	tgwallet "github.com/tonkeeper/tongo/wallet"
)

const (
	OpcodeStakeCoin = uint32(0x4cae3ab1)
)

type StakeInteractor struct {
	client             *liteapi.Client
	memoInteractor     *MemoInteractor
	contractInteractor *ContractInteractor
	stakeRepository    *repository.StakeRepository
	driverWallet       *tgwallet.Wallet
}

func NewStakeInteractor(client *liteapi.Client,
	memoInteractor *MemoInteractor,
	contractInteractor *ContractInteractor,
	stakeRepository *repository.StakeRepository,
	driverWallet *tgwallet.Wallet) *StakeInteractor {
	interactor := &StakeInteractor{
		client:             client,
		memoInteractor:     memoInteractor,
		contractInteractor: contractInteractor,
		stakeRepository:    stakeRepository,
		driverWallet:       driverWallet,
	}
	return interactor
}

func (interactor *StakeInteractor) Store(requests []domain.StakeRequest) error {
	for _, request := range requests {
		_, err := interactor.stakeRepository.InsertIfNotExists(request.Address, request.RoundSince, request.Hash, request.Info)
		if err != nil {
			log.Printf("🔴 inserting stake - %v\n", err.Error())
			return err
		}
	}
	return nil
}

func (interactor *StakeInteractor) LoadTriable() ([]domain.StakeRequest, error) {

	requests, err := interactor.stakeRepository.FindAllTriable(domain.GetMaxRetry())
	if err != nil {
		log.Printf("🔴 loading stake - %v\n", err.Error())
		return nil, err
	}

	return requests, nil
}

func (interactor *StakeInteractor) SendStakeMessageToJettonWallets(requests []domain.StakeRequest) error {
	// The round-since value must be considered as a condition whether to call stakeCoin or not.
	//
	//	start from sooner roundSince
	//  run get_method get_treasury_state
	//  extract participations field
	//  check if the desired round-since does exist in participations or not
	//  if exists, skip it, otherwise send stake-coin message

	// Split the wallets based on their round-since
	splitted := make(map[uint32][]domain.StakeRequest, 0)
	for _, request := range requests {
		roundSince := request.RoundSince
		if _, exist := splitted[roundSince]; exist {
			splitted[roundSince] = append(splitted[roundSince], request)
		} else {
			subList := make([]domain.StakeRequest, 0, 1)
			subList = append(subList, request)
			splitted[roundSince] = subList
		}
	}

	treasuryState, err := interactor.contractInteractor.GetTreasuryState()
	if err != nil {
		log.Printf("🔴 getting treasury state - %v\n", err.Error())
		return err
	}

	for roundSince, subList := range splitted {
		if _, exist := treasuryState.Participations[roundSince]; exist {
			continue
		}

		for _, request := range subList {
			accid, err := tongo.AccountIDFromBase64Url(request.Address)
			if err != nil {
				log.Printf("🔴 parsing wallet address %v - %v\n", request.Address, err.Error())
				continue
			}

			interactor.stakeRepository.SetRetrying(request.Address, roundSince, request.Hash, time.Now())

			// check the wallet to know if it is wating for a stake-coin messages, using get_wallet_state
			walletState, err := interactor.contractInteractor.GetWalletState(accid)
			if err != nil {
				log.Printf("🔴 getting wallet state - %v\n", err.Error())
				interactor.stakeRepository.SetState(request.Address, roundSince, request.Hash, domain.RequestStateError)
				continue
			}

			if _, exist := walletState.Staking[roundSince]; !exist {
				log.Printf("🔵 Wallet has no stake request.")
				interactor.stakeRepository.SetState(request.Address, roundSince, request.Hash, domain.RequestStateSkipped)
				continue
			}

			err = interactor.stakeCoin(accid, roundSince)
			if err != nil {
				log.Printf("🔴 staking [wallet: %v] - %v\n", request.Address, err.Error())
				interactor.stakeRepository.SetState(request.Address, roundSince, request.Hash, domain.RequestStateError)
				continue
			} else {
				interactor.stakeRepository.SetSuccess(request.Address, roundSince, request.Hash, time.Now())
				log.Printf("Successfully stakeed for wallet address %v.\n", request.Address)
			}
		}
	}

	return nil
}

func (interactor *StakeInteractor) stakeCoin(accid tongo.AccountID, roundSince uint32) error {
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
		log.Printf("🔴 sending Stake message - %v\n", err.Error())
		return err
	}

	return nil
}

func (interactor *StakeInteractor) MakeStakeRequests(trans []tongo.Transaction) []domain.StakeRequest {
	requests := make([]domain.StakeRequest, 0, 1)
	for _, t := range trans {
		ht := domain.NewHTransaction(&t.Transaction)

		// leave transaction if it's failed
		if !ht.IsSucceeded() {
			continue
		}

		msgs := ht.GetOutMessagesByOpcode(OpcodeSaveCoin)
		if len(msgs) > 1 {
			log.Printf("❗️ something's wrong, more than one msg found.")
			continue
		}

		var msg *domain.HMessage = nil
		if len(msgs) == 1 {
			msg = msgs[0]
		}

		info := domain.StakeRelatedInfo{
			Value: ht.Value(),
			Time:  ht.UnixTime(),
			Hash:  ht.Formatter().Hash(),
		}

		if msg != nil {
			accid := msg.Dest()
			cell := msg.GetBody()
			m := domain.SaveCoinMessage{}
			tlb.Unmarshal(cell, &m)

			addr := accid.ToHuman(true, domain.IsTestNet())
			requests = append(requests, domain.StakeRequest{
				Address:    addr,
				RoundSince: m.RoundSince,
				Hash:       ht.Formatter().Hash(),
				Info:       info,
				CreateTime: time.Now()})
		}
	}

	return requests
}

func findLastUnprocessed(trans []tongo.Transaction, lastHash string) int {
	for i, t := range trans {
		ht := domain.NewHTransaction(&t.Transaction)
		if ht.Formatter().Hash() == lastHash {
			return i
		}
	}

	return len(trans)
}
