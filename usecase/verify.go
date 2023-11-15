package usecase

import (
	"driver/domain"
	"driver/interface/exporter"
	"driver/interface/repository"
	"log"
	"math/big"
	"time"

	"github.com/tonkeeper/tongo"
	"github.com/tonkeeper/tongo/liteapi"
)

type VerifyInteractor struct {
	client             *liteapi.Client
	contractInteractor *ContractInteractor
	stakeRepository    *repository.StakeRepository
	unstakeRepository  *repository.UnstakeRepository
}

func NewVerifyInteractor(client *liteapi.Client,
	contractInteractor *ContractInteractor,
	stakeRepository *repository.StakeRepository,
	unstakeRepository *repository.UnstakeRepository) *VerifyInteractor {
	interactor := &VerifyInteractor{
		client:             client,
		contractInteractor: contractInteractor,
		stakeRepository:    stakeRepository,
		unstakeRepository:  unstakeRepository,
	}

	return interactor
}

func (interactor *VerifyInteractor) LoadVerifiable() ([]*domain.StakeRequest, []*domain.UnstakeRequest, error) {

	stakeRequests, err := interactor.stakeRepository.FindAllVerifiable()
	if err != nil {
		exporter.IncErrorCount()
		log.Printf("🔴 loading verifiable stake - %v\n", err.Error())
		return nil, nil, err
	}

	unstakeRequests, err := interactor.unstakeRepository.FindAllVerifiable()
	if err != nil {
		exporter.IncErrorCount()
		log.Printf("🔴 loading verifiable unstake - %v\n", err.Error())
		return nil, nil, err
	}

	return stakeRequests, unstakeRequests, nil
}

func (interactor *VerifyInteractor) VerifyStakeRequests(requests []*domain.StakeRequest) error {

	for _, request := range requests {
		log.Printf("verifying stake [wallet = %v]\n", request.Address)
		accid, err := tongo.AccountIDFromBase64Url(request.Address)
		if err != nil {
			exporter.IncErrorCount()
			log.Printf("🔴 verifying stake - parsing wallet address %v - %v\n", request.Address, err.Error())
			continue
		}

		// Check the wallet to know if it is wating for a stake-coin messages.
		walletState, err := interactor.contractInteractor.GetWalletState(accid)
		if err != nil {
			exporter.IncErrorCount()
			log.Printf("🔴 verifying stake - getting wallet state - %v\n", err.Error())
			continue
		}

		if _, exist := walletState.Staking[request.RoundSince]; !exist {
			// If it's not waiting for such this request, then we can assume the stake is done. So set
			// the state to 'verified'
			interactor.stakeRepository.SetVerified(request.Hash, time.Now())
		} else {
			// If the wallet is still waiting for such this request, it must filed to be done. So set
			// the state to 'retriable'.
			interactor.stakeRepository.SetState(request.Hash, domain.RequestStateRetriable)
		}
	}

	return nil
}

func (interactor *VerifyInteractor) VerifyUnstakeRequests(requests []*domain.UnstakeRequest) error {

	for _, request := range requests {
		log.Printf("verifying unstake [wallet = %v]\n", request.Address)
		accid, err := tongo.AccountIDFromBase64Url(request.Address)
		if err != nil {
			exporter.IncErrorCount()
			log.Printf("🔴 verifying unstake - parsing wallet address %v - %v\n", request.Address, err.Error())
			continue
		}

		// Check the wallet to know if it is wating for a withdraw messages.
		walletState, err := interactor.contractInteractor.GetWalletState(accid)
		if err != nil {
			exporter.IncErrorCount()
			log.Printf("🔴 verifying unstake - getting wallet state - %v\n", err.Error())
			continue
		}

		// Skip the request if the wallet has no unstaking
		if walletState.Unstaking.Cmp(big.NewInt(0)) == 0 {
			// If it's not waiting for any unstake request, then we can assume the unstake is done. So set
			// the state to 'verified'
			interactor.unstakeRepository.SetVerified(request.Hash, time.Now())
		} else {
			// If the wallet is still waiting for unstaking, it must filed to be done. So set
			// the state to 'retriable'.
			interactor.unstakeRepository.SetState(request.Hash, domain.RequestStateRetriable)
		}
	}

	return nil
}
