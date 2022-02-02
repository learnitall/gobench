package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var Verbose bool
var RunID string
var ElsaticsearchURL string
var ElasticsearchIndex string

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run a benchmark.",
	Long: `Set common options shared amongst benchmarks and then trigger
execution of a benchmark by name. Each subcommand represents a supported
benchmark which can be executed and has its own options.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
		os.Exit(1)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "Enables verbose debug info.")
	runCmd.PersistentFlags().StringVar(&RunID, "runid", "", "Set unique ID to identify benchmark results.")
	runCmd.PersistentFlags().StringVar(&ElsaticsearchURL, "elasticsearch-url", "", "Set URL of Elasticsearch instance to export results to.")
	runCmd.PersistentFlags().StringVar(&ElasticsearchIndex, "elasticsearch-index", "", "Set Elasticsearch Index to send results to.")
}
