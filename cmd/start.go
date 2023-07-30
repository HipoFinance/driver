/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"driver/domain"
	"driver/domain/config"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts driver's tasks",
	Long:  `Starts driver's tasks. To stop it, run 'stop' command.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("start called.")

		defaultDependencyInject()

		fmt.Printf("\n"+
			"----------------------------------\n"+
			"Network:             %s\n"+
			"Treasury wallet:     %v\n"+
			"Driver Interval:     %v\n"+
			"Extractoin Interval: %v\n"+
			"Stake Interval:      %v\n"+
			"Unstake Interval:    %v\n"+
			"----------------------------------\n",
			config.GetNetwork(),
			config.GetTreasuryAddress(),
			driverWallet.GetAddress().ToHuman(true, config.IsTestNet()),
			config.GetExtractInterval(),
			config.GetStakeInterval(),
			config.GetUnstakeInterval())

		config.GetTreasuryAddress()

		extractTiker := schedule(extract, config.GetExtractInterval(), quit)
		stakeTicker := schedule(stake, config.GetStakeInterval(), quit)
		unstakeTicker := schedule(unstake, config.GetUnstakeInterval(), quit)

		signal.Ignore()
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
		s := <-stop
		log.Printf("Got signal '%v', stopping", s)

		extractTiker.Stop()
		stakeTicker.Stop()
		unstakeTicker.Stop()
	},
}

func schedule(task func(), interval time.Duration, done chan bool) *time.Ticker {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {

			case <-ticker.C:
				ticker.Stop()
				task()
				ticker.Reset(interval)

			case <-done:
				return
			}
		}
	}()
	return ticker
}

func extract() {
	accountId := config.GetTreasuryAccountId()

	extractResult, err := extractInteractor.Extract(accountId)
	if err != nil {
		fmt.Printf("❌ No request is extracted: %v", err.Error())
		return
	}

	err = extractInteractor.Store(extractResult)
	if err != nil {
		fmt.Printf("❌ No request is stored: %v", err.Error())
		return
	}

	printOutWallets(extractResult)
}

func stake() {
	requests, err := stakeInteractor.LoadTriable()
	if err != nil {
		fmt.Printf("❌ Failed to send Stake messages - %v\n", err.Error())
		return
	}

	stakeInteractor.SendStakeMessageToJettonWallets(requests)
}

func unstake() {
	requests, err := unstakeInteractor.LoadTriable()
	if err != nil {
		fmt.Printf("❌ Failed to send Withdraw message - %v\n", err.Error())
		return
	}

	unstakeInteractor.SendWithdrawMessageToJettonWallets(requests)
}

func printOutWallets(extractResult *domain.ExtractionResult) {

	if len(extractResult.StakeRequests)+len(extractResult.UnstakeRequests) > 0 {
		fmt.Printf("------------- FOUND WALLET LIST -----------------\n")
	}

	i := 1
	for _, wallet := range extractResult.StakeRequests {
		info := wallet.Info
		fmt.Printf("stake #%03d - [ wallet: %v , hash: %v ]\n", i, wallet.Address, info.Hash)
		i++
	}

	for _, wallet := range extractResult.UnstakeRequests {
		info := wallet.Info
		fmt.Printf("unstake #%03d - [ wallet: %v , hash: %v ]\n", i, wallet.Address, info.Hash)
		i++
	}
}

func init() {
	rootCmd.AddCommand(startCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// startCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
