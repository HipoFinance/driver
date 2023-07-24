package domain

import (
	"time"

	"github.com/tonkeeper/tongo/tlb"
)

// @TOCLEAR: If round-since is a time, why it is not uint64?
type JettonWallet struct {
	Address    string                 `json:"address"`
	RoundSince uint32                 `json:"round_since"`
	Info       RelatedTransactionInfo `json:"info"`
	CreateTime time.Time              `json:"create_time"`
	NotifyTime *time.Time             `json:"notify_time"`
}

type RelatedTransactionInfo struct {
	Value tlb.Grams `json:"value"`
	Time  time.Time `json:"time"`
	Hash  string    `json:"hash"`
}
