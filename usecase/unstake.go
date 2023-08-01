package usecase

import (
	"context"
	"driver/domain"
	"driver/domain/config"
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
			log.Printf("🔴 inserting unstake - %v\n", err.Error())
			return err
		}
	}
	return nil
}

func (interactor *UnstakeInteractor) LoadTriable() ([]domain.UnstakeRequest, error) {

	requests, err := interactor.unstakeRepository.FindAllTriable(config.GetMaxRetry())
	if err != nil {
		log.Printf("🔴 loading unstake - %v\n", err.Error())
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

		accid, err := tongo.AccountIDFromBase64Url(request.Address)
		if err != nil {
			log.Printf("🔴 parsing wallet address %v - %v\n", request.Address, err.Error())
			continue
		}

		// Get maximum burnable tokens as the total budget for unstaking.
		totalBudget, err := interactor.contractInteractor.GetMaxBurnableTokens()
		if err != nil {
			log.Printf("🔴 getting max burnable tokens - %v\n", err.Error())
			continue
		}

		// The budget must be grater than request's tokens. If not, neither this request nor the next ones canbe payed,
		// because the list is sorted based on unstaking value.
		if request.Tokens.Cmp(totalBudget) > 0 {
			log.Printf("🔵 unstaking [wallet: %v] - requested tokens' value is more than available budget\n", request.Address)
			break
		}

		// Check the wallet to know if it is wating for a withdraw messages.
		walletState, err := interactor.contractInteractor.GetWalletState(accid)
		if err != nil {
			log.Printf("🔴 getting wallet state - %v\n", err.Error())
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
		if walletState.Unstaking.Cmp(totalBudget) > 0 {
			log.Printf("🔵 unstaking [wallet: %v] - postponed due to not enough budget\n", request.Address)
			continue
		}

		interactor.unstakeRepository.SetRetrying(request.Address, request.Tokens, request.Hash, time.Now())

		err = interactor.withdraw(accid, request.Tokens)
		if err != nil {
			log.Printf("🔴 unstaking [wallet: %v] - %v\n", request.Address, err.Error())
			interactor.unstakeRepository.SetState(request.Address, request.Tokens, request.Hash, domain.RequestStateError)
			continue
		} else {
			interactor.unstakeRepository.SetSuccess(request.Address, request.Tokens, request.Hash, time.Now())
			log.Printf("unstaking done [wallet: %v]\n", request.Address)
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
		log.Printf("🔴 sending withdraw [wallet: %v] - %v\n", accid.ToHuman(true, config.IsTestNet()), err.Error())
		return err
	}

	return nil
}

func (interactor *UnstakeInteractor) MakeUnstakeRequests(trans []tongo.Transaction) []domain.UnstakeRequest {
	requests := make([]domain.UnstakeRequest, 0, 1)
	for _, t := range trans {
		ht := domain.NewHTransaction(&t.Transaction)

		// Leave transaction if it's failed.
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
			err := tlb.Unmarshal(cell, &m)
			if err != nil {
				log.Printf("🔴 unmarshaling message body [trans hash: %v] - %v\n", ht.Formatter().Hash(), err.Error())
				continue
			}

			// @TODO: Use a better conversion method
			buff, err := m.Tokens.MarshalJSON()
			if err != nil {
				log.Printf("🔴 parsing tokens [value: %v] - %v\n", m.Tokens, err.Error())
				continue
			}
			buff = buff[1 : len(buff)-1] // remove " marks from begining and end of json value
			var tokens big.Int
			tokens.UnmarshalText(buff)
			addr := accid.ToHuman(true, config.IsTestNet())
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
