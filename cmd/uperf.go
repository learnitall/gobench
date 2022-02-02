/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"github.com/spf13/cobra"
)

// uperfCmd represents the uperf command
var uperfCmd = &cobra.Command{
	Use:   "uperf workload",
	Short: "Run the uperf networking benchmark.",
	Long:  `Uperf requires an xml file to define the workloads to run. This must be provided as the positional argument "workload".`,
	Args:  cobra.ExactValidArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func init() {
	runCmd.AddCommand(uperfCmd)
}
