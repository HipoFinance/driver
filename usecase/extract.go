package usecase

import (
	"context"
	"driver/domain"
	"fmt"
	"log"

	"github.com/tonkeeper/tongo"
	"github.com/tonkeeper/tongo/liteapi"
	tgwallet "github.com/tonkeeper/tongo/wallet"
)

const (
	OpcodeSaveCoin     = uint32(0x7f30ee55)
	OpcodeReserveToken = uint32(0x7bdd97de)
)

type ExtractInteractor struct {
	client             *liteapi.Client
	memoInteractor     *MemoInteractor
	contractInteractor *ContractInteractor
	stakeInteractor    *StakeInteractor
	unstakeInteractor  *UnstakeInteractor
	driverWallet       *tgwallet.Wallet
}

func NewExtractInteractor(client *liteapi.Client,
	memoInteractor *MemoInteractor,
	contractInteractor *ContractInteractor,
	stakeInteractor *StakeInteractor,
	unstakeInteractor *UnstakeInteractor,
	driverWallet *tgwallet.Wallet) *ExtractInteractor {
	interactor := &ExtractInteractor{
		client:             client,
		memoInteractor:     memoInteractor,
		contractInteractor: contractInteractor,
		stakeInteractor:    stakeInteractor,
		unstakeInteractor:  unstakeInteractor,
		driverWallet:       driverWallet,
	}
	return interactor
}

func (interactor *ExtractInteractor) Extract(treasuryAccount tongo.AccountID) (*domain.ExtractionResult, error) {

	result := domain.ExtractionResult{
		StakeRequests:   make([]domain.StakeRequest, 0, 50),
		UnstakeRequests: make([]domain.UnstakeRequest, 0, 50),
	}

	// Read the latest processed transaction info
	latestProcessedHash, err := interactor.memoInteractor.GetLatestProcessedHash()
	if err != nil {
		fmt.Printf("🔴 getting last processed hash - %v\n", err.Error())
		return nil, err
	}

	// Get last transactions of the treasury account, sorted decently by time.
	trans, err := interactor.client.GetLastTransactions(context.Background(), treasuryAccount, 50)
	if err != nil {
		fmt.Printf("🔴 getting last transactions - %v\n", err.Error())
		return nil, err
	}

	// Keep the first processing transaction's hash, so that in next call, stop searching through the processed transactions.
	var firstTrans *domain.HTransaction
	firstTransHash := ""
	if len(trans) > 0 {
		firstTrans = domain.NewHTransaction(&trans[0].Transaction)
		firstTransHash = firstTrans.Formatter().Hash()
	}

	// Start processing transactions
	var hash tongo.Bits256
	reachEnd := firstTransHash == latestProcessedHash
	if reachEnd {
		log.Printf("No new transaction for process.\n")
	}

	for err == nil && len(trans) > 0 && !reachEnd {
		index := findLastUnprocessed(trans, latestProcessedHash)
		reachEnd = index < len(trans)
		if reachEnd {
			trans = trans[0:index]
		}

		stkReqs := interactor.stakeInteractor.MakeStakeRequests(trans)
		unstkReqs := interactor.unstakeInteractor.MakeUnstakeRequests(trans)
		result.StakeRequests = append(result.StakeRequests, stkReqs...)
		result.UnstakeRequests = append(result.UnstakeRequests, unstkReqs...)
		log.Printf("Processing transactions... Total: %v / Found: %v stake(s) and %v unstake(s)\n", len(trans), len(stkReqs), len(unstkReqs))

		// If the latest processed transaction is not reached,
		if !reachEnd {
			// Extract the Lt and the Hash of last transaction
			lt := trans[len(trans)-1].Lt
			hash.FromHex(trans[len(trans)-1].Hash().Hex())

			// Extract previous transactions. GetTransactions function returns 16 items by max, this is why 16 is passed for the count parameter.
			trans, err = interactor.client.GetTransactions(context.Background(), 16, treasuryAccount, lt, hash)
			if err != nil {
				log.Printf("🔴 getting transactions - %v\n", err.Error())
				fmt.Printf("❌ No wallet will be kept due to above error.\n")
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
			log.Printf("🔴 updating latest hash - %v\n", err.Error())
		}
	}

	return &result, nil
}

func (interactor *ExtractInteractor) Store(extractResult *domain.ExtractionResult) error {
	err := interactor.stakeInteractor.Store(extractResult.StakeRequests)
	if err != nil {
		log.Printf("🔴 storing stake - %v\n", err.Error())
		return err
	}

	err = interactor.unstakeInteractor.Store(extractResult.UnstakeRequests)
	if err != nil {
		log.Printf("🔴 storing unstake - %v\n", err.Error())
		return err
	}

	return nil
}
