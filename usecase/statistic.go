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

	//##########################################################
	// tongo:
	// func ToBlockId(s tlb.ShardDesc, workchain int32) BlockIDExt
	// func ParseBlockID(s string) (BlockID, error)
	// func MustParseBlockID(s string) BlockID
	// func (s ShardID) MatchBlockID(block BlockID) bool

	// liteapi:
	// func (c *Client) targetBlock(ctx context.Context) (tongo.BlockIDExt, error)
	// func (c *Client) LookupBlock(ctx context.Context, blockID tongo.BlockID, mode uint32, lt *uint64, utime *uint32) (tongo.BlockIDExt, tlb.BlockInfo, error)
	// func (c *Client) GetShardInfo( ctx context.Context, blockID tongo.BlockIDExt, workchain uint32, shard uint64, exact bool) (tongo.BlockIDExt, error)
	// func (c *Client) GetAllShardsInfo(ctx context.Context, blockID tongo.BlockIDExt) ([]tongo.BlockIDExt, error)
	//##########################################################

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
