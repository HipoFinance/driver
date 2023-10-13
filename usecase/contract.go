package usecase

import (
	"context"
	"driver/domain/config"
	"driver/domain/model"
	"fmt"
	"log"
	"math/big"

	"github.com/tonkeeper/tongo"
	"github.com/tonkeeper/tongo/boc"
	"github.com/tonkeeper/tongo/liteapi"
	"github.com/tonkeeper/tongo/tlb"
)

var (
	ErrorUnexpectedTreasuryState = fmt.Errorf("unexpected treasury state")
	ErrorUnexpectedMaxBurnable   = fmt.Errorf("unexpected max burnable")
	ErrorUnexpectedWalletState   = fmt.Errorf("unexpected wallet state")
)

type ContractInteractor struct {
	client *liteapi.Client
}

func NewContractInteractor(client *liteapi.Client) *ContractInteractor {
	return &ContractInteractor{
		client: client,
	}
}

func (interactor *ContractInteractor) GetTreasuryState() (*model.TreasuryState, error) {
	code, stack, err := interactor.client.RunSmcMethod(context.Background(), config.GetTreasuryAccountId(), "get_treasury_state", tlb.VmStack{})

	if err != nil {
		log.Printf("🔴 getting treasury state [code = %v] - %v\n", code, err.Error())
		return nil, err
	}

	if len(stack) != 17 ||
		(stack[0].SumType != "VmStkTinyInt" && stack[0].SumType != "VmStkInt") ||
		(stack[1].SumType != "VmStkTinyInt" && stack[1].SumType != "VmStkInt") ||
		(stack[2].SumType != "VmStkTinyInt" && stack[2].SumType != "VmStkInt") ||
		(stack[3].SumType != "VmStkTinyInt" && stack[3].SumType != "VmStkInt") ||
		(stack[4].SumType != "VmStkTinyInt" && stack[4].SumType != "VmStkInt") ||
		(stack[5].SumType != "VmStkCell" && stack[5].SumType != "VmStkNull") ||
		(stack[6].SumType != "VmStkTinyInt") ||
		(stack[7].SumType != "VmStkTinyInt") ||
		(stack[8].SumType != "VmStkCell") ||
		(stack[9].SumType != "VmStkCell") ||
		(stack[10].SumType != "VmStkSlice") ||
		(stack[11].SumType != "VmStkSlice") ||
		(stack[12].SumType != "VmStkSlice") ||
		(stack[13].SumType != "VmStkSlice" && stack[13].SumType != "VmStkNull") ||
		stack[14].SumType != "VmStkTinyInt" ||
		(stack[15].SumType != "VmStkCell" && stack[15].SumType != "VmStkNull") ||
		(stack[16].SumType != "VmStkCell") {
		return nil, ErrorUnexpectedTreasuryState
	}

	result := &model.TreasuryState{}

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

	// result.BalancedRounds = stack[6].VmStkTinyInt != 0
	result.RoundsImbalance = uint8(stack[6].VmStkTinyInt)
	result.Stopped = stack[7].VmStkTinyInt != 0
	result.RewardShare = stack[14].VmStkTinyInt

	return result, nil
}

func (interactor *ContractInteractor) GetMaxBurnableTokens() (*big.Int, error) {
	code, stack, err := interactor.client.RunSmcMethod(context.Background(), config.GetTreasuryAccountId(), "get_max_burnable_tokens", tlb.VmStack{})

	if err != nil {
		log.Printf("🔴 getting max burnable tokens [code = %v] - %v\n", code, err.Error())
		return nil, err
	}

	if len(stack) != 1 ||
		(stack[0].SumType != "VmStkTinyInt" && stack[0].SumType != "VmStkInt") {
		return nil, ErrorUnexpectedMaxBurnable
	}

	result := getBigIntValue(stack[0], 0)

	return result, nil
}

func (interactor *ContractInteractor) GetWalletState(accountId tongo.AccountID) (*model.WalletState, error) {
	code, stack, err := interactor.client.RunSmcMethod(context.Background(), accountId, "get_wallet_state", tlb.VmStack{})

	if err != nil {
		log.Printf("🔴 getting wallet state [code = %v] - %v\n", code, err.Error())
		return nil, err
	}

	if len(stack) != 3 ||
		(stack[0].SumType != "VmStkTinyInt" && stack[0].SumType != "VmStkInt") ||
		(stack[1].SumType != "VmStkCell" && stack[1].SumType != "VmStkNull") ||
		(stack[2].SumType != "VmStkTinyInt" && stack[2].SumType != "VmStkInt") {
		return nil, ErrorUnexpectedWalletState
	}

	result := &model.WalletState{}

	result.Tokens.Set(getBigIntValue(stack[0], 0))
	result.Unstaking.Set(getBigIntValue(stack[2], 0))

	if stack[1].SumType == "VmStkCell" {
		result.Staking = make(map[uint32]tlb.Any)
		cell := stack[1].VmStkCell.Value
		var x tlb.Hashmap[tlb.Uint32, tlb.Any]
		x.UnmarshalTLB(&cell, tlb.NewDecoder())

		for i, key := range x.Keys() {
			result.Staking[uint32(key)] = x.Values()[i]
		}
	}

	return result, nil
}
func (interactor *ContractInteractor) GetTreasuryBalance() (uint64, error) {
	state, err := interactor.client.GetAccountState(context.Background(), config.GetTreasuryAccountId())
	if err != nil {
		return 0, err
	}

	return uint64(state.Account.Account.Storage.Balance.Grams), nil
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
