package usecase

import (
	"driver/domain"
	"driver/domain/config"
	"driver/domain/model"
	"driver/interface/repository"
	"log"
	"time"

	"github.com/tonkeeper/tongo"
	"github.com/tonkeeper/tongo/liteapi"
	"github.com/tonkeeper/tongo/tlb"
	tgwallet "github.com/tonkeeper/tongo/wallet"
)

type StakeInteractor struct {
	client             *liteapi.Client
	memoInteractor     *MemoInteractor
	contractInteractor *ContractInteractor
	stakeRepository    *repository.StakeRepository
	driverWallet       *tgwallet.Wallet

	messengerCh chan domain.MessagePack
	resposeCh   chan Response
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

func (interactor *StakeInteractor) InitializeChannel(messengerCh chan domain.MessagePack) chan Response {

	interactor.messengerCh = messengerCh

	interactor.resposeCh = make(chan Response, 5)
	go interactor.ListenOnResponse(interactor.resposeCh)
	return interactor.resposeCh
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

func (interactor *StakeInteractor) LoadTriable() ([]*domain.StakeRequest, error) {

	requests, err := interactor.stakeRepository.FindAllTriable(config.GetMaxRetry())
	if err != nil {
		log.Printf("🔴 loading stake - %v\n", err.Error())
		return nil, err
	}

	return requests, nil
}

func (interactor *StakeInteractor) SendStakeMessageToJettonWallets(requests []*domain.StakeRequest) error {
	// The round-since value must be considered as a condition whether to call stakeCoin or not.
	//
	//	start from sooner roundSince
	//  run get_method get_treasury_state
	//  extract participations field
	//  check if the desired round-since does exist in participations or not
	//  if exists, skip it, otherwise send stake-coin message

	// Split the wallets based on their round-since
	splitted := make(map[uint32][]*domain.StakeRequest, 0)
	for _, request := range requests {
		roundSince := request.RoundSince
		if _, exist := splitted[roundSince]; exist {
			splitted[roundSince] = append(splitted[roundSince], request)
		} else {
			subList := make([]*domain.StakeRequest, 0, 1)
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

			interactor.stakeRepository.SetRetrying(request.Hash, time.Now())

			// check the wallet to know if it is wating for a stake-coin messages, using get_wallet_state
			walletState, err := interactor.contractInteractor.GetWalletState(accid)
			if err != nil {
				log.Printf("🔴 getting wallet state - %v\n", err.Error())
				interactor.stakeRepository.SetState(request.Hash, domain.RequestStateError)
				continue
			}

			if _, exist := walletState.Staking[roundSince]; !exist {
				log.Printf("🔵 wallet has no stake request.")
				interactor.stakeRepository.SetState(request.Hash, domain.RequestStateSkipped)
				continue
			}

			reference := request.Hash
			mp := domain.MessagePack{
				Reference: reference,
				Message:   interactor.makeMessage(accid, request),

				StakeRequest:   request,
				UnstakeRequest: nil,
			}

			interactor.messengerCh <- mp
		}
	}

	return nil
}

func (interactor *StakeInteractor) makeMessage(accid tongo.AccountID, request *domain.StakeRequest) domain.Messagable {

	return domain.StakeCoinMessage{
		AccountId: accid,
		TlbMsg: domain.TlbStakeCoinMessage{
			Opcode:       tlb.Uint32(domain.OpcodeStakeCoin),
			QuieryId:     tlb.Uint64(time.Now().Unix()),
			RoundSince:   tlb.Uint32(request.RoundSince),
			ReturnExcess: tlb.MsgAddress{SumType: "AddrNone"},
		},
	}
}

func (interactor *StakeInteractor) MakeStakeRequests(trans []tongo.Transaction) []domain.StakeRequest {
	requests := make([]domain.StakeRequest, 0, 1)
	for _, t := range trans {
		ht := model.NewHTransaction(&t.Transaction)

		// leave transaction if it's failed
		if !ht.IsSucceeded() {
			continue
		}

		msgs := ht.GetOutMessagesByOpcode(OpcodeSaveCoin)
		if len(msgs) > 1 {
			log.Printf("❗️ something's wrong, more than one msg found.")
			continue
		}

		var msg *model.HMessage = nil
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
			tlbm := domain.TlbSaveCoinMessage{}
			tlb.Unmarshal(cell, &tlbm)

			addr := accid.ToHuman(true, config.IsTestNet())
			requests = append(requests, domain.StakeRequest{
				Address:    addr,
				RoundSince: uint32(tlbm.RoundSince),
				Hash:       ht.Formatter().Hash(),
				Info:       info,
				CreatedAt:  time.Now()})
		}
	}

	return requests
}

func (interactor *StakeInteractor) ListenOnResponse(respCh chan Response) {
	// @TODO: implement a way to end the loop and close the channel
	for {
		resp := <-respCh

		request := resp.StakeRequest
		if request == nil {
			log.Printf("🔴 staking [hash: %v] - request is nil!\n", resp.reference)
			continue
		}

		if !resp.ok {
			log.Printf("🔴 staking [wallet: %v] - %v\n", request.Address, resp.err.Error())
			interactor.stakeRepository.SetState(request.Hash, domain.RequestStateError)
		} else {
			interactor.stakeRepository.SetSent(request.Hash, time.Now())
			log.Printf("staking sent [wallet: %v]\n", request.Address)
		}
	}
}

func findLastUnprocessed(trans []tongo.Transaction, lastHash string) int {
	for i, t := range trans {
		ht := model.NewHTransaction(&t.Transaction)
		if ht.Formatter().Hash() == lastHash {
			return i
		}
	}

	return len(trans)
}
