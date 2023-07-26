package domain

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

type SaveCoinMessage struct {
	Opcode       uint32
	QuieryId     uint64
	Amount       tlb.Grams
	RoundSince   uint32
	ReturnExcess tlb.MsgAddress
}

type WithdrawMessage struct {
	Opcode       uint32
	QuieryId     uint64
	Tokens       big.Int
	Owner        tlb.MsgAddress
	ReturnExcess tlb.MsgAddress
}
