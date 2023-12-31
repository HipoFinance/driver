package model

import (
	"driver/domain/config"
	"driver/domain/util"
	"fmt"
	"log"
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

func (t *HTransaction) GetInMessagesByOpcode(opcode uint32) *HMessage {
	msg := t.InMessage()
	op := msg.Opcode()
	if opcode == op {
		return msg
	}

	return nil
}

func (t *HTransaction) GetOutMessagesByOpcode(opcode uint32) []*HMessage {
	res := make([]*HMessage, 0, 5)

	oMsgs := t.OutMessages()
	for _, msg := range oMsgs {
		op := msg.Opcode()
		if opcode == op {
			res = append(res, msg)
		}
	}

	return res
}

func (t *HTransaction) Formatter() *HTransactionFormatter {
	return NewHTransactionFormatter(t)
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
	return f.obj.Hash().Hex()
}

func (f *HTransactionFormatter) AccountId(format string) string {
	res := ""
	accid := f.obj.AccountId()
	switch strings.ToLower(format) {
	case AddrFormatRaw:
		res = accid.ToRaw()
	case AddrFormatBouncable:
		res = accid.ToHuman(true, config.IsTestNet())
	case AddrFormatNonBouncable:
		res = accid.ToHuman(false, config.IsTestNet())
	}
	return res
}

func (f *HTransactionFormatter) Src() string {
	addr := f.obj.InMessage().Src()
	return fmt.Sprintf("%v", addr)
}

func (f *HTransactionFormatter) Dest() string {
	addr := f.obj.InMessage().Dest()
	return fmt.Sprintf("%v", addr)
}

func (f *HTransactionFormatter) LocalTimeString() string {
	return f.obj.UnixTime().Local().Format(time.RFC1123)
}

func (f *HTransactionFormatter) ValueInGram() string {
	return util.GramString(int64(f.obj.Value()))
}

func (f *HTransactionFormatter) ValueInTon() string {
	return util.GramToTonString(int64(f.obj.Value()))
}

func LogTransaction(tx tlb.Transaction) {

	htx := NewHTransaction(&tx)

	accountAddress := htx.Formatter().AccountId(AddrFormatBouncable)
	hash := htx.Formatter().Hash()
	tf, _, _, _, _ := htx.Fees()

	totalFees := util.GramToTonString(int64(tf))
	lt := tx.Lt
	origStatus := tx.OrigStatus
	endStatus := tx.EndStatus
	now := time.Unix(int64(tx.Now), 0)

	msg := htx.InMessage()

	if msg != nil {
		src := msg.Src()
		msgSender := "-"
		if src != nil {
			msgSender = src.String()
		}

		dest := msg.Dest()
		msgDest := "-"
		if dest != nil {
			msgDest = dest.String()
		}

		msgOpcode := msg.Opcode()

		log.Printf("============================================================\n"+
			"In-Message\n"+
			"Account Address: %v\n"+
			"LT:              %v\n"+
			"Hash:            %v\n"+
			"Original Status: %v\n"+
			"End Status:      %v\n"+
			"Value:           %v\n"+
			"Total Fees:      %v\n"+
			"Now:             %v\n"+
			"Msg Sender:      %v\n"+
			"Msg Destination: %v\n"+
			"Msg Opcode:      %x\n",
			accountAddress, lt, hash, origStatus, endStatus, htx.Formatter().ValueInTon(), totalFees, now.Local().Format(time.RFC1123),
			msgSender, msgDest, msgOpcode)
	}

	outMsgs := htx.OutMessages()
	if len(outMsgs) > 0 {
		for _, msg := range outMsgs {
			src := msg.Src()
			msgSender := "-"
			if src != nil {
				msgSender = src.String()
			}

			dest := msg.Dest()
			msgDest := "-"
			if dest != nil {
				msgDest = dest.String()
			}
			msgOpcode := msg.Opcode()

			log.Printf("------------------------------------------------------------\n"+
				"Account Address: %v\n"+
				"LT:              %v\n"+
				"Hash:            %v\n"+
				"Original Status: %v\n"+
				"End Status:      %v\n"+
				"Value:           %v\n"+
				"Total Fees:      %v\n"+
				"Now:             %v\n"+
				"Msg Sender:      %v\n"+
				"Msg Destination: %v\n"+
				"Msg Opcode:      %v\n",
				accountAddress, lt, hash, origStatus, endStatus, htx.Formatter().ValueInTon(), totalFees, now.Local().Format(time.RFC1123),
				msgSender, msgDest, msgOpcode)

		}
	}
}
