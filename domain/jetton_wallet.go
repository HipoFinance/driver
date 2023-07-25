package domain

import (
	"time"

	"github.com/tonkeeper/tongo/tlb"
)

const (
	JWalletStateNew     = "new"
	JWalletStateOngoing = "ongoing"
	JWalletStateDone    = "done"
	JWalletStateSkipped = "skipped"
	JWalletStateError   = "error"
)

type JettonWallet struct {
	Address    string                 `json:"address"`
	RoundSince uint32                 `json:"round_since"`
	Hash       string                 `json:"hash"`
	State      string                 `json:"state"`
	Info       RelatedTransactionInfo `json:"info"`
	CreateTime time.Time              `json:"create_time"`
	NotifyTime *time.Time             `json:"notify_time"`
}

type RelatedTransactionInfo struct {
	Value tlb.Grams `json:"value"`
	Time  time.Time `json:"time"`
	Hash  string    `json:"hash"`
}
