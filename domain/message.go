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

func (m *HMessage) Opcode() []byte {
	body, _ := m.msg.Body.Value.MarshalJSON()
	cell := boc.NewCell()
	cell.UnmarshalJSON(body)
	opcode, _ := cell.ReadBytes(4)
	return opcode
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

func (t *HMessage) AllDestAddress() map[string]*tongo.AccountID {
	res := make(map[string]*tongo.AccountID, 10)
	// res["int/src"] = t.SrcInt()
	res["int/dest"] = t.DestInt()
	// res["ext-in/src"] = t.SrcExtIn()
	res["ext-in/dest"] = t.DestExtIn()
	// res["ext-out/src"] = t.SrcExtOut()
	res["ext-out/dest"] = t.DestExtOut()

	for key, msg := range res {
		if msg == nil {
			delete(res, key)
		}
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
