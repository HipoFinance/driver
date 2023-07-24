package domain

import "github.com/tonkeeper/tongo/tlb"

type SaveCoinMessage struct {
	Opcode       uint32
	QuieryId     uint64
	Amount       tlb.Grams
	RoundSince   uint32
	ReturnExcess tlb.MsgAddress
}
