package usecase

import (
	"context"
	"driver/domain"
	"log"
	"math/big"

	"github.com/tonkeeper/tongo/boc"
	"github.com/tonkeeper/tongo/liteapi"
	"github.com/tonkeeper/tongo/tlb"
)

type ContractInteractor struct {
	client *liteapi.Client
}

func NewContractInteractor(client *liteapi.Client) *ContractInteractor {
	return &ContractInteractor{
		client: client,
	}
}

func (interactor *ContractInteractor) GetTreasuryState() (*domain.TreasuryState, error) {
	code, stack, err := interactor.client.RunSmcMethod(context.Background(), domain.GetTreasuryAccountId(), "get_treasury_state", tlb.VmStack{})

	if err != nil {
		log.Printf("Failed to get treasury state [code = %v] - %v\n", code, err.Error())
	}

	// @TOCLEAR: Which one of the parameters can be null?
	// @TODO: use big.int for money values
	if len(stack) != 16 ||
		(stack[0].SumType != "VmStkTinyInt" && stack[0].SumType != "VmStkInt") ||
		(stack[1].SumType != "VmStkTinyInt" && stack[1].SumType != "VmStkInt") ||
		(stack[2].SumType != "VmStkTinyInt" && stack[2].SumType != "VmStkInt") ||
		(stack[3].SumType != "VmStkTinyInt" && stack[3].SumType != "VmStkInt") ||
		(stack[4].SumType != "VmStkTinyInt" && stack[4].SumType != "VmStkInt") ||
		(stack[5].SumType != "VmStkCell" && stack[5].SumType != "VmStkNull") ||
		stack[6].SumType != "VmStkTinyInt" ||
		(stack[7].SumType != "VmStkCell" && stack[7].SumType != "VmStkNull") ||
		(stack[8].SumType != "VmStkCell" && stack[8].SumType != "VmStkNull") ||
		(stack[9].SumType != "VmStkSlice" && stack[9].SumType != "VmStkNull") ||
		(stack[10].SumType != "VmStkSlice" && stack[10].SumType != "VmStkNull") ||
		(stack[11].SumType != "VmStkSlice" && stack[11].SumType != "VmStkNull") ||
		(stack[12].SumType != "VmStkSlice" && stack[12].SumType != "VmStkNull") ||
		stack[13].SumType != "VmStkTinyInt" ||
		(stack[14].SumType != "VmStkCell" && stack[14].SumType != "VmStkNull") ||
		(stack[15].SumType != "VmStkCell" && stack[15].SumType != "VmStkNull") {
		return nil, domain.ErrorInvalidTreasuryState
	}

	result := &domain.TreasuryState{}

	result.TotalCoins.Set(getBigIntValue(stack[0], 0))
	result.TotalTokens.Set(getBigIntValue(stack[1], 0))
	result.TotalStaking.Set(getBigIntValue(stack[2], 0))
	result.TotalUnstaking.Set(getBigIntValue(stack[3], 0))
	result.TotalValidatorStake.Set(getBigIntValue(stack[4], 0))

	if stack[5].SumType == "VmStkCell" {
		result.Participations = make(map[uint32]tlb.Any)
		cell := stack[5].VmStkCell.Value
		var x tlb.Hashmap[tlb.Uint32, tlb.Any]
		x.UnmarshalTLB(&cell, tlb.NewDecoder())

		for i, key := range x.Keys() {
			result.Participations[uint32(key)] = x.Values()[i]
		}
	}

	result.Stopped = stack[6].VmStkTinyInt != 0
	result.RewardShare = stack[13].VmStkTinyInt

	return result, nil
}

func getBigIntValue(stackItem tlb.VmStackValue, defaultValue int64) *big.Int {
	if stackItem.SumType == "VmStkTinyInt" {
		return big.NewInt(stackItem.VmStkTinyInt)
	}

	if stackItem.SumType == "VmStkInt" {
		var cell boc.Cell

		i257 := stackItem.VmStkInt
		i257.MarshalTLB(&cell, nil)

		var bi tlb.Int257
		bi.UnmarshalTLB(&cell, nil)

		result := big.Int(bi)
		return &result
	}

	return nil
}
