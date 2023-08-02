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
		tongoClient, err = liteapi.NewClientWithDefaultMainnet()
	case config.TestNetwork:
		tongoClient, err = liteapi.NewClientWithDefaultTestnet()
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
	statisticInteractor = usecase.NewStatisticInteractor(tongoClient)

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
var statisticInteractor *usecase.StatisticInteractor
var messengerInteractor *usecase.MessengerInteractor
var driverWallet wallet.Wallet
