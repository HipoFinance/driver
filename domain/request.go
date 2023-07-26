package domain

import (
	"math/big"
	"time"

	"github.com/tonkeeper/tongo/tlb"
)

const (
	RequestStateNew     = "new"
	RequestStateOngoing = "ongoing"
	RequestStateDone    = "done"
	RequestStateSkipped = "skipped"
	RequestStateError   = "error"
)

type StakeRequest struct {
	Address     string           `json:"address"`
	RoundSince  uint32           `json:"round_since"`
	Hash        string           `json:"hash"`
	State       string           `json:"state"`
	Retried     int              `json:"retried"`
	Info        StakeRelatedInfo `json:"info"`
	CreateTime  time.Time        `json:"create_time"`
	RetryTime   *time.Time       `json:"retry_time"`
	SuccessTime *time.Time       `json:"success_time"`
}

type StakeRelatedInfo struct {
	Value tlb.Grams `json:"value"`
	Time  time.Time `json:"time"`
	Hash  string    `json:"hash"`
}

type UnstakeRequest struct {
	Address     string           `json:"address"`
	Tokens      big.Int          `json:"tokens"`
	Hash        string           `json:"hash"`
	State       string           `json:"state"`
	Retried     int              `json:"retried"`
	Info        StakeRelatedInfo `json:"info"`
	CreateTime  time.Time        `json:"create_time"`
	RetryTime   *time.Time       `json:"retry_time"`
	SuccessTime *time.Time       `json:"success_time"`
}

type UnstakeRelatedInfo struct {
	Value tlb.Grams `json:"value"`
	Time  time.Time `json:"time"`
	Hash  string    `json:"hash"`
}
