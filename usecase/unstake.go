package usecase

import (
	"context"
	"driver/domain"
	"driver/interface/repository"
	"log"
	"math/big"
	"sort"
	"time"

	"github.com/tonkeeper/tongo"
	"github.com/tonkeeper/tongo/boc"
	"github.com/tonkeeper/tongo/liteapi"
	"github.com/tonkeeper/tongo/tlb"
	tgwallet "github.com/tonkeeper/tongo/wallet"
)

const (
	OpcodeWithdraw = uint32(0x469bd91e)
)

type UnstakeInteractor struct {
	client             *liteapi.Client
	memoInteractor     *MemoInteractor
	contractInteractor *ContractInteractor
	unstakeRepository  *repository.UnstakeRepository
	driverWallet       *tgwallet.Wallet
}

func NewUnstakeInteractor(client *liteapi.Client,
	memoInteractor *MemoInteractor,
	contractInteractor *ContractInteractor,
	unstakeRepository *repository.UnstakeRepository,
	driverWallet *tgwallet.Wallet) *UnstakeInteractor {
	interactor := &UnstakeInteractor{
		client:             client,
		memoInteractor:     memoInteractor,
		contractInteractor: contractInteractor,
		unstakeRepository:  unstakeRepository,
		driverWallet:       driverWallet,
	}
	return interactor
}

func (interactor *UnstakeInteractor) Store(requests []domain.UnstakeRequest) error {
	for _, request := range requests {
		_, err := interactor.unstakeRepository.InsertIfNotExists(request.Address, request.Tokens, request.Hash, request.Info)
		if err != nil {
			log.Printf("Failed to insert unstake record - %v\n", err.Error())
			return err
		}
	}
	return nil
}

func (interactor *UnstakeInteractor) LoadTriable() ([]domain.UnstakeRequest, error) {

	requests, err := interactor.unstakeRepository.FindAllTriable(domain.GetMaxRetry())
	if err != nil {
		log.Printf("Failed to load unstake records - %v\n", err.Error())
		return nil, err
	}

	return requests, nil
}

func (interactor *UnstakeInteractor) SendWithdrawMessageToJettonWallets(requests []domain.UnstakeRequest) error {

	// Sorts the requests based on Tokens value accending, so that most requests will be done with a specified budget.
	sort.Slice(requests, func(i, j int) bool {
		return requests[i].Tokens.Cmp(&requests[j].Tokens) < 0
	})

	for _, request := range requests {

		// Check the treaury treasuryBalance.
		treasuryBalance, err := interactor.contractInteractor.GetTreasuryBalance()
		if err != nil {
			log.Printf("Failed to get treasury balance - %v\n", err.Error())
			return err
		}

		accid, err := tongo.AccountIDFromBase64Url(request.Address)
		if err != nil {
			log.Printf("Failed to parse wallet address %v - %v\n", request.Address, err.Error())
			continue
		}

		// Treasury budget must be grater than request's tokens. If not, break the loop as the next request's tokens
		// are more than this one (because the list are sorted).
		// @TODO: Use get_method to find out the available budget
		var totalBudget big.Int
		totalBudget.SetUint64(treasuryBalance - 10)
		if request.Tokens.Cmp(&totalBudget) > 0 {
			break
		}

		// @TODO: check the wallet to know if it is wating for a withdraw messages, using get_wallet_state
		walletState, err := interactor.contractInteractor.GetWalletState(accid)
		if err != nil {
			log.Printf("Failed to get wallet state - %v\n", err.Error())
			interactor.unstakeRepository.SetState(request.Address, request.Tokens, request.Hash, domain.RequestStateError)
			continue
		}

		// Skip the request if the wallet has no unstaking
		if walletState.Unstaking.Cmp(big.NewInt(0)) == 0 {
			log.Printf("No request for unstaking %v\n", request.Address)
			interactor.unstakeRepository.SetState(request.Address, request.Tokens, request.Hash, domain.RequestStateSkipped)
			continue
		}

		// Check if the budget can pay the required unstaking value.
		// Note that each wallet may have multiple unstake requests. If so, the wallet keep the total as unstaking value.
		// So, compare the unstaking value of the wallet against the treasury budget, not the request.Token value.
		if walletState.Unstaking.Cmp(&totalBudget) > 0 {
			log.Printf("Not enough budget for unstaking %v\n", request.Address)
			continue
		}

		interactor.unstakeRepository.SetRetrying(request.Address, request.Tokens, request.Hash, time.Now())

		err = interactor.withdraw(accid, request.Tokens)
		if err != nil {
			log.Printf("Failed to unstake coin for wallet address %v - %v\n", request.Address, err.Error())
			interactor.unstakeRepository.SetState(request.Address, request.Tokens, request.Hash, domain.RequestStateError)
			continue
		} else {
			interactor.unstakeRepository.SetSuccess(request.Address, request.Tokens, request.Hash, time.Now())
			log.Printf("Successfully unstake coin for wallet address %v.\n", request.Address)
		}
	}

	return nil
}

func (interactor *UnstakeInteractor) withdraw(accid tongo.AccountID, tokens big.Int) error {
	queryId := uint64(time.Now().Unix())

	// @TODO: complete the cell data
	cell := boc.NewCell()
	cell.WriteUint(uint64(OpcodeWithdraw), 32) // opcode
	cell.WriteUint(queryId, 64)                // query id
	cell.WriteUint(0, 2)                       // return excess

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

func (interactor *UnstakeInteractor) MakeUnstakeRequests(trans []tongo.Transaction) []domain.UnstakeRequest {
	requests := make([]domain.UnstakeRequest, 0, 1)
	for _, t := range trans {
		ht := domain.NewHTransaction(&t.Transaction)

		// leave transaction if it's failed
		if !ht.IsSucceeded() {
			continue
		}

		msg := ht.GetInMessagesByOpcode(OpcodeReserveToken)

		info := domain.UnstakeRelatedInfo{
			Value: ht.Value(),
			Time:  ht.UnixTime(),
			Hash:  ht.Formatter().Hash(),
		}

		if msg != nil {
			accid := msg.Src()
			cell := msg.GetBody()
			m := domain.ReserveTokenMessage{}
			tlb.Unmarshal(cell, &m)

			// @TODO: Use a better conversion method
			buff, err := m.Tokens.MarshalJSON()
			if err != nil {
				log.Printf("Failed to parse tokens value: %v\n", err.Error())
				continue
			}
			buff = buff[1 : len(buff)-1] // remove " marks from begining and end of json value
			var tokens big.Int
			tokens.UnmarshalText(buff)
			addr := accid.ToHuman(true, domain.IsTestNet())
			requests = append(requests, domain.UnstakeRequest{
				Address:    addr,
				Tokens:     tokens,
				Hash:       ht.Formatter().Hash(),
				Info:       info,
				CreateTime: time.Now()})
		}
	}

	return requests
}
