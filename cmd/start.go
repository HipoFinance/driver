/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"driver/domain"
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

		extractTiker := schedule(extract, domain.GetExtractInterval(), quit)
		stakeTicker := schedule(stake, domain.GetStakeInterval(), quit)
		unstakeTicker := schedule(unstake, domain.GetUnstakeInterval(), quit)

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
	accountId := domain.GetTreasuryAccountId()

	extractResult, err := extractInteractor.Extract(accountId)
	if err != nil {
		fmt.Printf("❌ No stake is extracted due to error: %v", err.Error())
		return
	}

	err = extractInteractor.Store(extractResult)
	if err != nil {
		fmt.Printf("❌ No stake is stored due to error: %v", err.Error())
		return
	}

	printOutWallets(extractResult)
}

func stake() {
	wallets, err := stakeInteractor.LoadTriable()
	if err != nil {
		fmt.Printf("❌ Failed to Send stake message to jetton wallets - %v\n", err.Error())
		return
	}

	stakeInteractor.SendStakeMessageToJettonWallets(wallets)
}

func unstake() {
	wallets, err := unstakeInteractor.LoadTriable()
	if err != nil {
		fmt.Printf("❌ Failed to Send withdraw message to jetton wallets - %v\n", err.Error())
		return
	}

	unstakeInteractor.SendWithdrawMessageToJettonWallets(wallets)
}

func printOutWallets(extractResult *domain.ExtractionResult) {

	fmt.Printf("------------- FOUND WALLET LIST -----------------\n")
	i := 1
	for _, wallet := range extractResult.StakeRequests {
		fmt.Printf("stake #%03d - %v [ ", i, wallet.Address)
		sep := ""
		info := wallet.Info
		fmt.Printf("%v%v", sep, info.Hash)
		fmt.Printf(" ]\n")
		i++
	}

	for _, wallet := range extractResult.UnstakeRequests {
		fmt.Printf("unstake #%03d - %v [ ", i, wallet.Address)
		sep := ""
		info := wallet.Info
		fmt.Printf("%v%v", sep, info.Hash)
		fmt.Printf(" ]\n")
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
