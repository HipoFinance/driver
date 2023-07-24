package domain

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/tonkeeper/tongo"
	"github.com/tonkeeper/tongo/tlb"
)

const (
	AddrFormatRaw          = "raw"
	AddrFormatBouncable    = "bouncable"
	AddrFormatNonBouncable = "non-bouncable"
)

type HTransaction struct {
	trans *tlb.Transaction
}

func NewHTransaction(trans *tlb.Transaction) *HTransaction {
	return &HTransaction{
		trans: trans,
	}
}

func (t *HTransaction) Hash() tongo.Bits256 {
	b256 := t.trans.Hash()
	return tongo.Bits256(b256)
}

func (t *HTransaction) AccountId() *tongo.AccountID {
	b256 := t.trans.AccountAddr
	buf := tongo.Bits256(b256)
	return tongo.NewAccountId(0, buf)
}

func (t *HTransaction) Lt() uint64 {
	return t.trans.Lt
}

func (t *HTransaction) UnixTime() time.Time {
	return time.Unix(int64(t.trans.Now), 0)
}

func (t *HTransaction) Value() tlb.Grams {
	var value tlb.Grams

	value = 0
	if t.trans.Msgs.InMsg.Value.Value.Info.IntMsgInfo != nil {
		value = t.trans.Msgs.InMsg.Value.Value.Info.IntMsgInfo.Value.Grams
	}

	return value
}

func (t *HTransaction) InMessage() *HMessage {
	msg := t.trans.Msgs.InMsg.Value.Value
	return &HMessage{
		msg: &msg,
	}
}

func (t *HTransaction) OutMessages() []*HMessage {
	msgs := make([]*HMessage, t.trans.OutMsgCnt)

	ms := t.trans.Msgs.OutMsgs.Values()
	for i, m := range ms {
		msg := m.Value
		msgs[i] = NewHMessage(&msg)
	}

	return msgs
}

func (t *HTransaction) Fees() (totalFee, processFee, storageFee, inMsgFee, outMsgsFee tlb.Grams) {

	// Storage and process fees

	storageFee = t.trans.Description.TransOrd.StoragePh.Value.StorageFeesCollected
	processFee = t.trans.TotalFees.Grams - storageFee

	// Input message fees

	inMsg := t.InMessage()

	importFee := inMsg.ImportFee()
	fwdFee := inMsg.FwdFee()
	ihrFee := inMsg.IhrFee()

	inMsgFee = importFee + fwdFee + ihrFee

	// Output messages fees

	outMsgs := t.OutMessages()
	outMsgsFee = tlb.Grams(0)
	for _, msg := range outMsgs {

		importFee = msg.ImportFee()
		fwdFee = msg.FwdFee()
		ihrFee = msg.IhrFee()

		outMsgsFee += importFee + fwdFee + ihrFee
	}

	totalFee = storageFee + processFee + inMsgFee + outMsgsFee

	return
}

func (t *HTransaction) IsSucceeded() bool {
	return t.trans.Description.TransOrd.Action.Value.Value.Success
}

func (t *HTransaction) GetMessagesHavingOpcode(opcode uint32) []*HMessage {
	res := make([]*HMessage, 0, 5)

	msg := t.InMessage()
	op := msg.Opcode()
	if opcode == op {
		res = append(res, msg)
	}

	oMsgs := t.OutMessages()
	for _, msg := range oMsgs {
		op = msg.Opcode()
		if opcode == op {
			res = append(res, msg)
		}
	}

	return res
}

func (t *HTransaction) AllDestAddress() map[string]*tongo.AccountID {
	res := make(map[string]*tongo.AccountID, 10)

	subres := t.InMessage().AllDestAddress()
	for key, value := range subres {
		res["in/"+key] = value
	}

	for i, msg := range t.OutMessages() {
		idx := fmt.Sprintf("[%v]", i)
		subres := msg.AllDestAddress()
		for key, value := range subres {
			res["out"+idx+"/"+key] = value
		}
	}

	for key, msg := range res {
		if msg == nil {
			delete(res, key)
		}
	}
	return res
}

//---------------------------------

type HTransactionFormatter struct {
	// Output formatter
	obj *HTransaction
}

func NewHTransactionFormatter(obj *HTransaction) *HTransactionFormatter {
	return &HTransactionFormatter{
		obj: obj,
	}
}

func (f *HTransactionFormatter) Hash() string {
	buf := f.obj.Hash()
	return base64.URLEncoding.EncodeToString(buf[:])
}

func (f *HTransactionFormatter) AccountId(format string) string {
	res := ""
	acid := f.obj.AccountId()
	switch strings.ToLower(format) {
	case AddrFormatRaw:
		res = acid.ToRaw()
	case AddrFormatBouncable:
		res = acid.ToHuman(true, IsTestNet())
	case AddrFormatNonBouncable:
		res = acid.ToHuman(false, IsTestNet())
	}
	return res
}

func (f *HTransactionFormatter) Src() string {
	addr := f.obj.InMessage().SrcInt()
	return fmt.Sprintf("%v", addr)
}

func (f *HTransactionFormatter) Dest() string {
	addr := f.obj.InMessage().DestInt()
	return fmt.Sprintf("%v", addr)
}

func (f *HTransactionFormatter) LocalTimeString() string {
	return f.obj.UnixTime().Local().Format(time.RFC1123)
}

func (f *HTransactionFormatter) Value() string {
	return fmt.Sprintf("%v", f.obj.Value())
}
