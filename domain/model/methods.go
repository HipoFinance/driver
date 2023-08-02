package model

import (
	"math/big"

	"github.com/tonkeeper/tongo/tlb"
)

// @TODO: Read the commented-out fields
type TreasuryState struct {
	TotalCoins          big.Int
	TotalTokens         big.Int
	TotalStaking        big.Int
	TotalUnstaking      big.Int
	TotalValidatorStake big.Int
	Participations      map[uint32]tlb.Any
	Stopped             bool
	// WalletCode          tlb.Cell
	// LoanCode            tlb.Cell
	// Driver              tlb.Slice
	// Halter              tlb.Slice
	// Governor            tlb.Slice
	// ProposedGovernor    tlb.Cell
	RewardShare int64
	// RewardHistory       tlb.Cell
	// Content             tlb.Cell
}

type WalletState struct {
	Tokens    big.Int
	Staking   map[uint32]tlb.Any
	Unstaking big.Int
}
