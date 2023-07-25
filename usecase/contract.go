package usecase

import (
	"context"
	"driver/domain"
	"log"

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
	if len(stack) != 16 ||
		stack[0].SumType != "VmStkTinyInt" ||
		stack[1].SumType != "VmStkTinyInt" ||
		stack[2].SumType != "VmStkTinyInt" ||
		stack[3].SumType != "VmStkTinyInt" ||
		stack[4].SumType != "VmStkTinyInt" ||
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

	result.TotalCoins = stack[0].VmStkTinyInt
	result.TotalTokens = stack[1].VmStkTinyInt
	result.TotalStaking = stack[2].VmStkTinyInt
	result.TotalUnstaking = stack[3].VmStkTinyInt
	result.TotalValidatorStake = stack[4].VmStkTinyInt

	if stack[5].SumType == "VmStkCell" {
		result.Participations = make(map[uint32]string)
		cell := stack[5].VmStkCell.Value
		tlb.Unmarshal(&cell, &result.Participations)
	}

	result.Stopped = stack[6].VmStkTinyInt != 0
	result.RewardShare = stack[13].VmStkTinyInt

	return result, nil
}
