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

// startCmd represents the find command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts driver's tasks",
	Long:  `Starts driver's tasks. To stop it, run 'stop' command.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("start called.")

		defaultDependencyInject()

		findTiker := schedule(find, domain.GetFindInterval(), quit)
		stakeTicker := schedule(stake, domain.GetStakeInterval(), quit)
		unstakeTicker := schedule(unstake, domain.GetUnstakeInterval(), quit)

		signal.Ignore()
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
		s := <-stop
		log.Printf("Got signal '%v', stopping", s)

		findTiker.Stop()
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

func find() {
	accountId := domain.GetTreasuryAccountId()

	wallets, err := jettonWalletInteractor.FindJettonWallets(accountId)
	if err != nil {
		fmt.Printf("❌ No jetton wallet is extracted due to error: %v", err.Error())
		return
	}

	err = jettonWalletInteractor.Store(wallets)
	if err != nil {
		fmt.Printf("❌ No jetton wallet is stored due to error: %v", err.Error())
		return
	}

	printOutWallets(wallets)
}

func stake() {
	wallets, err := jettonWalletInteractor.LoadNotNotified()
	if err != nil {
		fmt.Printf("❌ Failed to Send message to jetton wallets - %v\n", err.Error())
		return
	}

	jettonWalletInteractor.SendMessageToJettonWallets(wallets)
}

func unstake() {
	// wallets, err := jettonWalletInteractor.LoadNotNotified()
	// if err != nil {
	// 	fmt.Printf("❌ Failed to Send message to jetton wallets - %v\n", err.Error())
	// 	return
	// }

	// jettonWalletInteractor.SendMessageToJettonWallets(wallets)
}

func printOutWallets(wallets map[string]domain.JettonWallet) {

	fmt.Printf("------------- FOUND WALLET LIST -----------------\n")
	max := 3
	i := 1
	for _, wallet := range wallets {
		fmt.Printf("#%03d - %v [ ", i, wallet.Address)
		sep := ""
		for j, info := range wallet.Info {
			fmt.Printf("%v%v", sep, info.Hash)
			if j >= (max-1) && len(wallet.Info) > max {
				fmt.Printf(" and %v more...", len(wallet.Info)-j)
				break
			} else if j < len(wallet.Info)-1 {
				sep = " , "
			} else {
				sep = ""
			}
		}
		fmt.Printf(" ]\n")
		i++
	}
}

func init() {
	rootCmd.AddCommand(startCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// findCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// findCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
