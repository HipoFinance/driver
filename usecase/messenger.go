package usecase

import (
	"context"
	"driver/domain"
	"fmt"
	"log"
	"time"

	"github.com/tonkeeper/tongo/liteapi"
	tgwallet "github.com/tonkeeper/tongo/wallet"
)

const (
	IssuerStake   = "stake"
	IssuerUnstake = "unstake"
)

var ErrorTimeOut = fmt.Errorf("timeout for new seqno")

type Response struct {
	reference string
	ok        bool
	err       error
}

type MessengerInteractor struct {
	client       *liteapi.Client
	driverWallet *tgwallet.Wallet
	listenerCh   chan domain.MessagePack
	stakeCh      chan Response
	unstakeCh    chan Response
}

func NewMessengerInteractor(client *liteapi.Client,
	driverWallet *tgwallet.Wallet,
	listenerCh chan domain.MessagePack,
	stakeCh chan Response,
	unstakeCh chan Response) *MessengerInteractor {
	interactor := &MessengerInteractor{
		client:       client,
		driverWallet: driverWallet,
		listenerCh:   listenerCh,
		stakeCh:      stakeCh,
		unstakeCh:    unstakeCh,
	}
	return interactor
}

func (interactor *MessengerInteractor) ListenOnChannel() error {

	var err error = nil
	var seqno uint32 = 0

	// @TODO: implement a way to end the loop and close the channel
	for {
		msg := <-interactor.listenerCh
		if msg.Reference == "-close-" {
			break
		}

		err = interactor.driverWallet.Send(context.Background(), msg.Message.MakeMessage())
		if err != nil {
			log.Printf("🔴 sending message [issuer: %v, reference: %v] - %v\n", msg.Issuer, msg.Reference, err.Error())
		} else {
			seqno, err = interactor.waitForNextSeqno(seqno)
		}

		response := Response{
			reference: msg.Reference,
			ok:        err == nil,
			err:       err,
		}

		switch msg.Issuer {
		case IssuerStake:
			interactor.stakeCh <- response
		case IssuerUnstake:
			interactor.unstakeCh <- response
		default:
			log.Printf("🔴 sending response - unknown issuer! [isuuer: %v]\n", msg.Issuer)
		}
	}

	return nil
}

func (interactor *MessengerInteractor) waitForNextSeqno(seqno uint32) (uint32, error) {
	driverAccountId := interactor.driverWallet.GetAddress()

	err := ErrorTimeOut
	currSeqno := seqno

	start := time.Now()
	for time.Now().Before(start.Add(30 * time.Second)) {
		currSeqno, err = interactor.client.GetSeqno(context.Background(), driverAccountId)
		if err != nil {
			log.Printf("🔴 getting current driver's seqno - %v\n", err.Error())
		}

		if currSeqno > seqno {
			err = nil
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	return currSeqno, err
}
