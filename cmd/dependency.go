package cmd

import (
	"database/sql"
	"driver/domain"
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
	dbURI := domain.GetDbUri()
	dbPool, err = sql.Open("postgres", dbURI)
	if err != nil {
		log.Fatal(err)
	}
	dbPool.SetMaxOpenConns(20)
	dbPool.SetMaxIdleConns(5)
	dbPool.SetConnMaxIdleTime(1 * time.Minute)
	dbPool.SetConnMaxLifetime(4 * time.Hour)

	dbHandler := dbhandler.DBHandler{DB: dbPool}

	switch strings.ToLower(domain.GetNetwork()) {
	case domain.MainNetwork:
		tongoClient, err = liteapi.NewClientWithDefaultMainnet()
	case domain.TestNetwork:
		tongoClient, err = liteapi.NewClientWithDefaultTestnet()
	default:
		fmt.Printf("⛔️ Configuration paramet 'network' must be either 'main' or 'test' only.")
		return
	}

	if err != nil {
		log.Fatal("Unable to create tongo client: ", err)
	}

	driverWallet, err = wallet.New(domain.GetDriverWalletPrivateKey(), wallet.V3R2, 0, nil, tongoClient)

	jettonWalletRepository := repository.NewJettonWalletRepository(dbHandler)

	jettonWalletInteractor = usecase.NewJettonWalletInteractor(tongoClient, jettonWalletRepository, &driverWallet)
}

var dbPool *sql.DB
var tongoClient *liteapi.Client
var jettonWalletInteractor *usecase.JettonWalletInteractor
var driverWallet wallet.Wallet
