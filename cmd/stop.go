/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// startCmd represents the find command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stops driver's tasks",
	Long:  `Stops driver's tasks, which are started previously by 'start' command.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("stop called.")

		// send an integer to the 'quit' channel, defined in 'start' command file.
		quit <- true
		close(quit)
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// findCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// findCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
