package domain

import (
	"github.com/tonkeeper/tongo"
	"github.com/tonkeeper/tongo/boc"
	"github.com/tonkeeper/tongo/tlb"
	"github.com/tonkeeper/tongo/wallet"
)

const (
	OpcodeStakeCoin = uint32(0x4cae3ab1)
	OpcodeWithdraw  = uint32(0x469bd91e)
)

type MessagePack struct {
	Reference string
	Message   Messagable

	StakeRequest   *StakeRequest
	UnstakeRequest *UnstakeRequest
}

type Messagable interface {
	MakeMessage() *wallet.Message
}

type TlbSaveCoinMessage struct {
	Opcode       tlb.Uint32
	QuieryId     tlb.Uint64
	StakeAmount  tlb.Grams
	RoundSince   tlb.Uint32
	ReturnExcess tlb.MsgAddress
}

type TlbReserveTokenMessage struct {
	Opcode       tlb.Uint32
	QuieryId     tlb.Uint64
	Tokens       tlb.Grams
	Owner        tlb.MsgAddress
	ReturnExcess tlb.MsgAddress
}

type TlbStakeCoinMessage struct {
	Opcode       tlb.Uint32
	QuieryId     tlb.Uint64
	RoundSince   tlb.Uint32
	ReturnExcess tlb.MsgAddress
}

type TlbWithdrawMessage struct {
	Opcode       tlb.Uint32
	QuieryId     tlb.Uint64
	ReturnExcess tlb.MsgAddress
}

type StakeCoinMessage struct {
	AccountId tongo.AccountID
	TlbMsg    TlbStakeCoinMessage
}

type WithdrawMessage struct {
	AccountId tongo.AccountID
	TlbMsg    TlbWithdrawMessage
}

func (msg StakeCoinMessage) MakeMessage() *wallet.Message {

	cell := boc.NewCell()
	tlb.Marshal(cell, msg.TlbMsg)

	wmsg := wallet.Message{
		Amount:  150000000,     //  tlb.Grams
		Address: msg.AccountId, //  tongo.AccountID
		Body:    cell,          //  *boc.Cell
		Code:    nil,           //  *boc.Cell
		Data:    nil,           //  *boc.Cell
		Bounce:  true,          //  bool
		Mode:    1,             //  uint8	/ Pay transfer fees separately from the message value /
	}

	return &wmsg
}

func (msg WithdrawMessage) MakeMessage() *wallet.Message {

	cell := boc.NewCell()
	tlb.Marshal(cell, msg.TlbMsg)

	wmsg := wallet.Message{
		Amount:  300000000,     //  tlb.Grams
		Address: msg.AccountId, //  tongo.AccountID
		Body:    cell,          //  *boc.Cell
		Code:    nil,           //  *boc.Cell
		Data:    nil,           //  *boc.Cell
		Bounce:  true,          //  bool
		Mode:    1,             //  uint8	/ Pay transfer fees separately from the message value /
	}

	return &wmsg
}
