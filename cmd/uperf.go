package cmd

import (
	"github.com/learnitall/gobench/benchmarks/uperf"
	"github.com/spf13/cobra"
)

func runUperf(cmd *cobra.Command, args []string) {
	uperfCmdArgs := []string{
		"uperf", "-m", args[0],
	}
	uperfCmdArgs = append(uperfCmdArgs, args[1:]...)

	uperf := &uperf.UperfBenchmark{
		WorkloadPath: args[0],
		Cmd:          uperfCmdArgs,
	}
	RunBenchmark(uperf)
}

// uperfCmd represents the uperf command
var uperfCmd = &cobra.Command{
	Use:   "uperf workload options ...",
	Short: "Run the uperf networking benchmark.",
	Long:  `Uperf requires an xml file to define the workloads to run. This must be provided as the positional argument "workload". If you would like to pass CLI arguments to uperf, place them after the workload filename.`,
	Args:  cobra.MinimumNArgs(2),
	Run:   runUperf,
}

func init() {
	runCmd.AddCommand(uperfCmd)
}
