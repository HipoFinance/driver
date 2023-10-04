package domain

import (
	"math/big"
	"time"

	"github.com/tonkeeper/tongo/tlb"
)

const (
	RequestStateNew       = "new"
	RequestStateOngoing   = "ongoing"
	RequestStateSent      = "sent"
	RequestStateVerified  = "verified"
	RequestStateRetriable = "retriable"
	RequestStateSkipped   = "skipped"
	RequestStateError     = "error"
)

type StakeRequest struct {
	Address    string           `json:"address"`
	RoundSince uint32           `json:"round_since"`
	Hash       string           `json:"hash"`
	State      string           `json:"state"`
	RetryCount int              `json:"retry_count"`
	Info       StakeRelatedInfo `json:"info"`
	CreatedAt  time.Time        `json:"created_at"`
	RetriedAt  *time.Time       `json:"retried_at"`
	SentAt     *time.Time       `json:"sent_at"`
	VerifiedAt *time.Time       `json:"verified_at"`
}

type StakeRelatedInfo struct {
	Value tlb.Grams `json:"value"`
	Time  time.Time `json:"time"`
	Hash  string    `json:"hash"`
}

type UnstakeRequest struct {
	Address    string             `json:"address"`
	Tokens     big.Int            `json:"tokens"`
	Hash       string             `json:"hash"`
	State      string             `json:"state"`
	RetryCount int                `json:"retry_count"`
	Info       UnstakeRelatedInfo `json:"info"`
	CreatedAt  time.Time          `json:"created_at"`
	RetriedAt  *time.Time         `json:"retried_at"`
	SentAt     *time.Time         `json:"sent_at"`
	VerifiedAt *time.Time         `json:"verified_at"`
}

type UnstakeRelatedInfo struct {
	Value tlb.Grams `json:"value"`
	Time  time.Time `json:"time"`
	Hash  string    `json:"hash"`
}

type ExtractionResult struct {
	StakeRequests   []StakeRequest
	UnstakeRequests []UnstakeRequest
}
