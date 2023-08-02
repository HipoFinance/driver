package domain

import (
	"time"

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
	Issuer    string
	Reference string
	Message   Messagable
}

type Messagable interface {
	MakeMessage() *wallet.Message
}

type SaveCoinMessage struct {
	AccountId    tongo.AccountID
	Opcode       uint32
	QuieryId     uint64
	Amount       tlb.Grams
	RoundSince   uint32
	ReturnExcess tlb.MsgAddress
}

type ReserveTokenMessage struct {
	AccountId    tongo.AccountID
	Opcode       uint32
	QuieryId     uint64
	Tokens       tlb.Grams
	Owner        tlb.MsgAddress
	ReturnExcess tlb.MsgAddress
}

func (msg SaveCoinMessage) MakeMessage() *wallet.Message {

	queryId := uint64(time.Now().Unix())

	cell := boc.NewCell()
	cell.WriteUint(uint64(OpcodeStakeCoin), 32) // opcode
	cell.WriteUint(queryId, 64)                 // query id
	cell.WriteUint(uint64(msg.RoundSince), 32)  // round since
	cell.WriteUint(0, 2)                        // return excess

	wmsg := wallet.Message{
		Amount:  100000000,     //  tlb.Grams
		Address: msg.AccountId, //  tongo.AccountID
		Body:    cell,          //  *boc.Cell
		Code:    nil,           //  *boc.Cell
		Data:    nil,           //  *boc.Cell
		Bounce:  true,          //  bool
		Mode:    1,             //  uint8	/ Pay transfer fees separately from the message value /
	}

	return &wmsg
}

func (msg ReserveTokenMessage) MakeMessage() *wallet.Message {

	queryId := uint64(time.Now().Unix())

	cell := boc.NewCell()
	cell.WriteUint(uint64(OpcodeWithdraw), 32) // opcode
	cell.WriteUint(queryId, 64)                // query id
	cell.WriteUint(0, 2)                       // return excess

	wmsg := wallet.Message{
		Amount:  100000000,     //  tlb.Grams
		Address: msg.AccountId, //  tongo.AccountID
		Body:    cell,          //  *boc.Cell
		Code:    nil,           //  *boc.Cell
		Data:    nil,           //  *boc.Cell
		Bounce:  true,          //  bool
		Mode:    1,             //  uint8	/ Pay transfer fees separately from the message value /
	}

	return &wmsg
}
