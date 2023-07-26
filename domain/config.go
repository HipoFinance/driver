package domain

import (
	"crypto/ed25519"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/viper"
	"github.com/tonkeeper/tongo"
	"github.com/tonkeeper/tongo/wallet"
)

const (
	MainNetwork = "mainnet"
	TestNetwork = "testnet"
)

var (
	ErrorInvalidNetwork = fmt.Errorf("network must be equal to 'mainnet' or 'testnet' only")

	ErrorNoMnemonic          = fmt.Errorf("no mnemonic is defined")
	ErrorMnemonicConflict    = fmt.Errorf("only one of mnemonic or mnemonic_url must be defined")
	ErrorReadingMnemonicFile = fmt.Errorf("error in reading mnemonic file")

	ErrorInvalidExtractInterval = fmt.Errorf("invalid time interval for extract process")
	ErrorInvalidStakeInterval   = fmt.Errorf("invalid time interval for stake process")
	ErrorInvalidUnstakeInterval = fmt.Errorf("invalid time interval for unstake process")

	ErrorInvalidTreausryAddress = fmt.Errorf("invalid treasury address")
)

var (
	TrailingSlashRE = regexp.MustCompile("/+$")
)

var (
	dbUri   string
	network string

	mnemonic               string
	mnemonic_url           string
	driverWalletPrivateKey ed25519.PrivateKey

	treasuryAddress   string
	treasuryAccountId tongo.AccountID

	extractInterval time.Duration
	stakeInterval   time.Duration
	unstakeInterval time.Duration
)

func ReadConfig(filePath string) {
	viper.SetConfigFile(filePath)

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("⚠️ Failed reading config file: %v\n", err.Error())
	}

	err := initializeVariables()
	if err != nil {
		log.Fatalf("Configuration error - %v\n", err.Error())
	}
}

// This method processes the configuration parameters and keeps the processed values
// in some variables for later accesses rapidly.
func initializeVariables() error {
	var err error

	// Database stuff
	dbUri = TrailingSlashRE.ReplaceAllString(viper.GetString("service_db_uri"), "")

	// Network stuff
	network = strings.TrimSpace(strings.ToLower(viper.GetString("network")))
	if strings.Compare(network, MainNetwork) != 0 && strings.Compare(network, TestNetwork) != 0 {
		return ErrorInvalidNetwork
	}

	// Treasury stuff
	treasuryAddress = strings.TrimSpace(viper.GetString("treasury_address"))
	treasuryAccountId, err = tongo.AccountIDFromBase64Url(treasuryAddress)
	if err != nil {
		return ErrorInvalidTreausryAddress
	}

	// Driver wallet stuff
	mnemonic = strings.TrimSpace(viper.GetString("mnemonic"))
	mnemonic_url = strings.TrimSpace(viper.GetString("mnemonic_url"))
	if mnemonic == "" && mnemonic_url == "" {
		return ErrorNoMnemonic
	}
	if mnemonic != "" && mnemonic_url != "" {
		return ErrorMnemonicConflict
	}

	seed := mnemonic
	if mnemonic_url != "" {
		seed, err = readMnemonicFile(mnemonic_url)
		if err != nil {
			return ErrorReadingMnemonicFile
		}
	}

	driverWalletPrivateKey, err = wallet.SeedToPrivateKey(seed)
	if err != nil {
		log.Printf("Failed to get private key - %v\n", err.Error())
		return err
	}

	//---------------------------------------------------------------
	// extract interval
	strValue := viper.GetString("extract_interval")
	extractInterval, err = time.ParseDuration(strValue)
	if err != nil {
		return ErrorInvalidExtractInterval
	}

	//---------------------------------------------------------------
	// stake interval
	strValue = viper.GetString("stake_interval")
	stakeInterval, err = time.ParseDuration(strValue)
	if err != nil {
		return ErrorInvalidStakeInterval
	}

	//---------------------------------------------------------------
	// unstake interval
	strValue = viper.GetString("unstake_interval")
	unstakeInterval, err = time.ParseDuration(strValue)
	if err != nil {
		return ErrorInvalidUnstakeInterval
	}

	return nil
}

func readMnemonicFile(filePath string) (string, error) {

	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Failed to read mmnemonic file - %v\n", err.Error())
		return "", err
	}

	// Convert []byte to string
	content := string(fileContent)
	return content, nil
}

//-------------------------------------------------------------------
// Processed values

//-------------------------------------------------------------------
// Normal configuration values

func GetDbUri() string {
	return dbUri
}

func GetTreasuryAddress() string {
	return treasuryAddress
}

func GetTreasuryAccountId() tongo.AccountID {
	return treasuryAccountId
}

func GetNetwork() string {
	return network
}

func GetExtractInterval() time.Duration {
	return extractInterval
}

func GetStakeInterval() time.Duration {
	return stakeInterval
}

func GetUnstakeInterval() time.Duration {
	return unstakeInterval
}

func GetDriverWalletPrivateKey() ed25519.PrivateKey {
	return driverWalletPrivateKey
}

// -------------------------------------------------------------------
// Evaluating values

func IsTestNet() bool {
	return strings.Compare(network, TestNetwork) == 0
}
