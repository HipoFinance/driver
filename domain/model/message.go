package model

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

func (m *HMessage) GetBody() *boc.Cell {
	body, _ := m.msg.Body.Value.MarshalJSON()
	cell := boc.NewCell()
	cell.UnmarshalJSON(body)
	return cell
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

func (m *HMessage) Src() *tongo.AccountID {
	src := m.srcInt()
	if src != nil {
		return src
	}

	src = m.srcExtIn()
	if src != nil {
		return src
	}
	src = m.srcExtOut()
	if src != nil {
		return src
	}

	return nil
}

func (m *HMessage) Dest() *tongo.AccountID {
	dest := m.destInt()
	if dest != nil {
		return dest
	}

	dest = m.destExtIn()
	if dest != nil {
		return dest
	}

	dest = m.destExtOut()
	if dest != nil {
		return dest
	}

	return nil
}

func (m *HMessage) srcInt() *tongo.AccountID {
	value := tlb.MsgAddress{}
	if m.msg.Info.IntMsgInfo != nil {
		value = m.msg.Info.IntMsgInfo.Src
	}
	acntId, _ := tongo.AccountIDFromTlb(value)
	return acntId
}

func (m *HMessage) destInt() *tongo.AccountID {
	value := tlb.MsgAddress{}
	if m.msg.Info.IntMsgInfo != nil {
		value = m.msg.Info.IntMsgInfo.Dest
	}
	acntId, _ := tongo.AccountIDFromTlb(value)
	return acntId
}

func (m *HMessage) srcExtIn() *tongo.AccountID {
	value := tlb.MsgAddress{}
	if m.msg.Info.ExtInMsgInfo != nil {
		value = m.msg.Info.ExtInMsgInfo.Src
	}
	acntId, _ := tongo.AccountIDFromTlb(value)
	return acntId
}

func (m *HMessage) destExtIn() *tongo.AccountID {
	value := tlb.MsgAddress{}
	if m.msg.Info.ExtInMsgInfo != nil {
		value = m.msg.Info.ExtInMsgInfo.Dest
	}
	acntId, _ := tongo.AccountIDFromTlb(value)
	return acntId
}

func (m *HMessage) srcExtOut() *tongo.AccountID {
	value := tlb.MsgAddress{}
	if m.msg.Info.ExtOutMsgInfo != nil {
		value = m.msg.Info.ExtOutMsgInfo.Src
	}
	acntId, _ := tongo.AccountIDFromTlb(value)
	return acntId
}

func (m *HMessage) destExtOut() *tongo.AccountID {
	value := tlb.MsgAddress{}
	if m.msg.Info.ExtOutMsgInfo != nil {
		value = m.msg.Info.ExtOutMsgInfo.Dest
	}
	acntId, _ := tongo.AccountIDFromTlb(value)
	return acntId
}

func (m *HMessage) Formatter() *HMessageFormatter {
	return NewHMessageFormatter(m)
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
