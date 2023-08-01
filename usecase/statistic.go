package usecase

import (
	"context"
	"driver/domain"
	"fmt"

	"github.com/tonkeeper/tongo"
	"github.com/tonkeeper/tongo/liteapi"
)

type StatisticInteractor struct {
	client *liteapi.Client
}

func NewStatisticInteractor(client *liteapi.Client) *StatisticInteractor {
	interactor := &StatisticInteractor{
		client: client,
	}
	return interactor
}

func (interactor *StatisticInteractor) Statistic(treasuryAccount tongo.AccountID) (*domain.StatisticResult, error) {

	result := domain.StatisticResult{}

	// var wc int32 = 0
	// var shard uint64 = 0
	// var seqno uint32 = 10000

	trans, err := interactor.client.GetLastTransactions(context.Background(), treasuryAccount, 100)
	if err != nil {
		return nil, err
	}
	tr := trans[len(trans)-1]
	blockId := tr.BlockID

	trids, ok, err := interactor.client.ListBlockTransactions(context.Background(), blockId, 7, 100, nil)
	if err != nil {
		return nil, err
	}

	if ok {
		for i, tr := range trids {
			fmt.Printf("#%v: %v\n", i, tr.Hash)

			// var cell boc.Cell
			// var bin tlb.Int256
			// bin.UnmarshalTLB(&cell, nil)
			// var acc tlb.Account
			// tlb.Unmarshal(&cell, acc)

			// interactor.client.GetTransactions(context.Background(), 100, acc, *tr.Lt, tongo.Bits256(*tr.Hash))
		}
	}

	return &result, nil
}

func (interactor *StatisticInteractor) Store(StatisticResult *domain.StatisticResult) error {

	return nil
}
