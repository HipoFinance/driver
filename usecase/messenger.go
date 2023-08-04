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

var ErrorTimeOut = fmt.Errorf("timeout for new seqno")

type Response struct {
	reference string
	ok        bool
	err       error

	StakeRequest   *domain.StakeRequest
	UnstakeRequest *domain.UnstakeRequest
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
			log.Printf("🔴 sending message [reference: %v] - %v\n", msg.Reference, err.Error())
		} else {
			seqno, err = interactor.waitForNextSeqno(seqno)
		}

		response := Response{
			reference: msg.Reference,
			ok:        err == nil,
			err:       err,

			StakeRequest:   msg.StakeRequest,
			UnstakeRequest: msg.UnstakeRequest,
		}

		if msg.StakeRequest != nil {
			// The message is a stake request, so send the response to stake response channel
			interactor.stakeCh <- response
		} else if msg.UnstakeRequest != nil {
			// The message is an unstake request, so send the response to unstake response channel
			interactor.unstakeCh <- response
		} else {
			// Oops! neither of request objects is provided
			log.Printf("🔴 sending response - unknown request source! [reference: %v]\n", msg.Reference)
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
