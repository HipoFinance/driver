package domain

import (
	"fmt"
	"math/big"

	"github.com/tonkeeper/tongo/tlb"
)

var (
	ErrorInvalidTreasuryState = fmt.Errorf("invalid treasury state")
)

// @TOCLEAR: What are the type/structure of the marked fields?
type TreasuryState struct {
	TotalCoins          big.Int
	TotalTokens         big.Int
	TotalStaking        big.Int
	TotalUnstaking      big.Int
	TotalValidatorStake big.Int
	Participations      map[uint32]tlb.Any
	Stopped             bool
	// WalletCode          string // ?
	// LoanCode            string // ?
	// Driver              string // ?
	// Halter              string // ?
	// Governor            string // ?
	// ProposedGovernor    string // ?
	RewardShare int64
	// RewardHistory       string // ?
	// Content             string // ?
}
