package domain

import "fmt"

var (
	ErrorInvalidTreasuryState = fmt.Errorf("invalid treasury state")
)

// @TOCLEAR: What are the type/structure of the marked fields?
type TreasuryState struct {
	TotalCoins          int64
	TotalTokens         int64
	TotalStaking        int64
	TotalUnstaking      int64
	TotalValidatorStake int64
	Participations      map[uint32]string // ?
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
