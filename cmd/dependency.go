package cmd

import (
	"database/sql"
	"driver/domain"
	"driver/domain/config"
	"driver/infrastructure/dbhandler"
	"driver/interface/repository"
	"driver/usecase"
	"fmt"
	"log"
	"strings"
	"time"

	tconfig "github.com/tonkeeper/tongo/config"
	"github.com/tonkeeper/tongo/liteapi"
	"github.com/tonkeeper/tongo/wallet"
)

func defaultDependencyInject() {
	var err error
	dbURI := config.GetDbUri()
	dbPool, err = sql.Open("postgres", dbURI)
	if err != nil {
		log.Fatal(err)
	}
	dbPool.SetMaxOpenConns(20)
	dbPool.SetMaxIdleConns(5)
	dbPool.SetConnMaxIdleTime(1 * time.Minute)
	dbPool.SetConnMaxLifetime(4 * time.Hour)

	dbHandler := dbhandler.DBHandler{DB: dbPool}

	switch strings.ToLower(config.GetNetwork()) {
	case config.MainNetwork:
		// tongoClient, err = liteapi.NewClientWithDefaultMainnet()
		servers := make([]tconfig.LiteServer, 0, 1)
		servers = append(servers, tconfig.LiteServer{
			Host: "65.21.233.98:34796",
			Key:  "5TTG6JCfkbscgAUksjKXFlEsMoNM2fm/3NK5w9jnFWc="
		})
		tongoClient, err = liteapi.NewClient(liteapi.WithLiteServers(servers))

	case config.TestNetwork:
		servers := make([]tconfig.LiteServer, 0, 4)
		servers = append(servers, tconfig.LiteServer{
			Host: "65.108.204.54:29296",
			Key:  "p2tSiaeSqX978BxE5zLxuTQM06WVDErf5/15QToxMYA=",
		})
		servers = append(servers, tconfig.LiteServer{
			Host: "178.63.63.122:20700",
			Key:  "dGLlRRai3K9FGkI0dhABmFHMv+92QEVrvmTrFf5fbqA=",
		})
		servers = append(servers, tconfig.LiteServer{
			Host: "116.202.225.189:20700",
			Key:  "24RL7iVI20qcG+j//URfd/XFeEG9qtezW2wqaYQgVKw=",
		})
		servers = append(servers, tconfig.LiteServer{
			Host: "65.108.141.177:17439",
			Key:  "0MIADpLH4VQn+INHfm0FxGiuZZAA8JfTujRqQugkkA8=",
		})
		tongoClient, err = liteapi.NewClient(liteapi.WithLiteServers(servers))
		// tongoClient, err = liteapi.NewClientWithDefaultTestnet()
	default:
		fmt.Printf("⛔️ Configuration paramet 'network' must be either 'main' or 'test' only.")
		return
	}

	if err != nil {
		log.Fatal("Unable to create tongo client: ", err)
	}

	driverWallet, err = wallet.New(config.GetDriverWalletPrivateKey(), wallet.V4R2, 0, nil, tongoClient)
	if err != nil {
		log.Fatalf("Unable to connect to driver wallet - %v\n", err.Error())
		return
	}

	stakeRepository := repository.NewStakeRepository(dbHandler)
	unstakeRepository := repository.NewUnstakeRepository(dbHandler)
	memoRepository := repository.NewMemoRepository(dbHandler)

	memoInteractor = usecase.NewMemoInteractor(memoRepository)
	contractInteractor = usecase.NewContractInteractor(tongoClient)
	stakeInteractor = usecase.NewStakeInteractor(tongoClient, memoInteractor, contractInteractor, stakeRepository, &driverWallet)
	unstakeInteractor = usecase.NewUnstakeInteractor(tongoClient, memoInteractor, contractInteractor, unstakeRepository, &driverWallet)
	extractInteractor = usecase.NewExtractInteractor(tongoClient, memoInteractor, contractInteractor, stakeInteractor, unstakeInteractor, &driverWallet)
	verifyInteractor = usecase.NewVerifyInteractor(tongoClient, contractInteractor, stakeRepository, unstakeRepository)

	messengerCh := make(chan domain.MessagePack, 10)
	stakeCh := stakeInteractor.InitializeChannel(messengerCh)
	unstakeCh := unstakeInteractor.InitializeChannel(messengerCh)

	messengerInteractor = usecase.NewMessengerInteractor(tongoClient, &driverWallet, messengerCh, stakeCh, unstakeCh)
}

var dbPool *sql.DB
var tongoClient *liteapi.Client

var memoInteractor *usecase.MemoInteractor
var contractInteractor *usecase.ContractInteractor
var stakeInteractor *usecase.StakeInteractor
var unstakeInteractor *usecase.UnstakeInteractor
var extractInteractor *usecase.ExtractInteractor
var verifyInteractor *usecase.VerifyInteractor

var messengerInteractor *usecase.MessengerInteractor
var driverWallet wallet.Wallet
