package domain

import (
	"time"

	"github.com/tonkeeper/tongo/tlb"
)

type JettonWallet struct {
	Address    string                 `json:"address"`
	RoundSince uint32                 `json:"round_since"`
	MsgHash    string                 `json:"msg_hash"`
	Info       RelatedTransactionInfo `json:"info"`
	CreateTime time.Time              `json:"create_time"`
	NotifyTime *time.Time             `json:"notify_time"`
}

type RelatedTransactionInfo struct {
	Value tlb.Grams `json:"value"`
	Time  time.Time `json:"time"`
	Hash  string    `json:"hash"`
}
