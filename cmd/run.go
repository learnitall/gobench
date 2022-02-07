package cmd

import (
	"os"

	"github.com/learnitall/gobench/define"

	"github.com/spf13/cobra"
)

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
	cfg := define.GetConfig()
	runCmd.PersistentFlags().BoolVarP(&cfg.Verbose, "verbose", "v", false, "Enables verbose debug info.")
	runCmd.PersistentFlags().StringVarP(&cfg.RunID, "uuid", "u", "", "Set unique run UUID ID to identify benchmark results.")
	runCmd.PersistentFlags().StringVar(&cfg.ElasticsearchURL, "elasticsearch-url", "", "Set URL of Elasticsearch instance to export results to.")
	runCmd.PersistentFlags().StringVar(&cfg.ElasticsearchIndex, "elasticsearch-index", "", "Set Elasticsearch Index to send results to.")
}
