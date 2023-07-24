package domain

import (
	"github.com/tonkeeper/tongo"
	"github.com/tonkeeper/tongo/boc"
	"github.com/tonkeeper/tongo/tlb"
)

type HMessage struct {
	msg *tlb.Message
}

func NewHMessage(msg *tlb.Message) *HMessage {
	return &HMessage{
		msg: msg,
	}
}

func (m *HMessage) Opcode() uint32 {
	body, _ := m.msg.Body.Value.MarshalJSON()
	cell := boc.NewCell()
	cell.UnmarshalJSON(body)
	opcode, _ := cell.ReadUint(32)
	return uint32(opcode)
}

func (m *HMessage) ImportFee() tlb.Grams {
	var fee tlb.Grams
	fee = 0
	if m.msg.Info.ExtInMsgInfo != nil {
		fee = m.msg.Info.ExtInMsgInfo.ImportFee
	}
	return fee
}

func (m *HMessage) FwdFee() tlb.Grams {
	var fee tlb.Grams
	fee = 0
	if m.msg.Info.IntMsgInfo != nil {
		fee = m.msg.Info.IntMsgInfo.FwdFee
	}
	return fee
}

func (m *HMessage) IhrFee() tlb.Grams {
	var fee tlb.Grams
	fee = 0
	if m.msg.Info.IntMsgInfo != nil {
		fee = m.msg.Info.IntMsgInfo.IhrFee
	}
	return fee
}

func (m *HMessage) Value() tlb.Grams {
	var value tlb.Grams
	value = 0
	if m.msg.Info.IntMsgInfo != nil {
		value = m.msg.Info.IntMsgInfo.Value.Grams
	}
	return value
}

func (m *HMessage) SrcInt() *tongo.AccountID {
	value := tlb.MsgAddress{}
	if m.msg.Info.IntMsgInfo != nil {
		value = m.msg.Info.IntMsgInfo.Src
	}
	acntId, _ := tongo.AccountIDFromTlb(value)
	return acntId
}

func (m *HMessage) DestInt() *tongo.AccountID {
	value := tlb.MsgAddress{}
	if m.msg.Info.IntMsgInfo != nil {
		value = m.msg.Info.IntMsgInfo.Dest
	}
	acntId, _ := tongo.AccountIDFromTlb(value)
	return acntId
}

func (m *HMessage) SrcExtIn() *tongo.AccountID {
	value := tlb.MsgAddress{}
	if m.msg.Info.ExtInMsgInfo != nil {
		value = m.msg.Info.ExtInMsgInfo.Src
	}
	acntId, _ := tongo.AccountIDFromTlb(value)
	return acntId
}

func (m *HMessage) DestExtIn() *tongo.AccountID {
	value := tlb.MsgAddress{}
	if m.msg.Info.ExtInMsgInfo != nil {
		value = m.msg.Info.ExtInMsgInfo.Dest
	}
	acntId, _ := tongo.AccountIDFromTlb(value)
	return acntId
}

func (m *HMessage) SrcExtOut() *tongo.AccountID {
	value := tlb.MsgAddress{}
	if m.msg.Info.ExtOutMsgInfo != nil {
		value = m.msg.Info.ExtOutMsgInfo.Src
	}
	acntId, _ := tongo.AccountIDFromTlb(value)
	return acntId
}

func (m *HMessage) DestExtOut() *tongo.AccountID {
	value := tlb.MsgAddress{}
	if m.msg.Info.ExtOutMsgInfo != nil {
		value = m.msg.Info.ExtOutMsgInfo.Dest
	}
	acntId, _ := tongo.AccountIDFromTlb(value)
	return acntId
}

// @TODO: Check if there is only one destination address for any message, return one AccountId rather than an array
func (m *HMessage) AllDestAddress() []*tongo.AccountID {
	res := make([]*tongo.AccountID, 0, 1)

	accid := m.DestInt()
	if accid != nil {
		res = append(res, accid)
	}

	accid = m.DestExtIn()
	if accid != nil {
		res = append(res, accid)
	}

	accid = m.DestExtOut()
	if accid != nil {
		res = append(res, accid)
	}

	return res
}

//---------------------------------

type HMessageFormatter struct {
	// Output formatter
	obj *HMessage
}

func NewHMessageFormatter(obj *HMessage) *HMessageFormatter {
	return &HMessageFormatter{
		obj: obj,
	}
}
