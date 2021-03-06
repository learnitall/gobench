package cmd

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/learnitall/gobench/define"
	"github.com/learnitall/gobench/exporters"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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
	runCmd.PersistentFlags().BoolVarP(&cfg.Quiet, "quiet", "q", false, "Disable all log output. Overrides the --verbose/-v.")
	runCmd.PersistentFlags().BoolVarP(&cfg.PrintJson, "print-json", "p", false, "Print benchmark results as json documents. Guaranteed that the printed data is jq-pipeable, i.e. gobench run --quiet --print-json ... | jq.")
	runCmd.PersistentFlags().StringVarP(&cfg.RunID, "uuid", "u", uuid.New().String(), "Set unique run UUID ID to identify benchmark results. If one is not given, one will be generated.")
	runCmd.PersistentFlags().StringVar(&cfg.ElasticsearchURL, "elasticsearch-url", "", "Set URL of Elasticsearch instance to export results to.")
	runCmd.PersistentFlags().StringVar(&cfg.ElasticsearchIndex, "elasticsearch-index", "", "Set Elasticsearch Index to send results to.")
	runCmd.PersistentFlags().BoolVar(&cfg.ElasticsearchInjectProductHeader,
		"elasticsearch-iph", true, `Have the Elasticsearch http client inject
'X-Product-Elastic=ElasticSearch' header into ElasticSearch server responses.`)
}

// SetLogLevel sets the current log level based on the given Config.
func SetLogLevel(config *define.Config) {
	if config.Quiet {
		zerolog.SetGlobalLevel(zerolog.Disabled)
	} else if config.Verbose {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}

// LogVersion tries to log the current version of gobench at the info level.
func LogVersion() {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		log.Warn().Msg("Failed to read build info, unable to get version.")
		return
	}
	log.Info().
		Str("version", bi.Main.Version).
		Msg("Starting gobench.")
}

// GetExporter takes in the current Config and returns an appropriate Exporterable.
func GetExporter(config *define.Config) define.Exporterable {
	configuredExporters := []define.Exporterable{}
	if config.ElasticsearchURL != "" {
		log.Info().Msg("Creating ElasticsearchExporter.")
		configuredExporters = append(
			configuredExporters, &exporters.ElasticsearchExporter{},
		)
	}
	if config.PrintJson {
		log.Info().Msg("Creating JsonExporter.")
		configuredExporters = append(
			configuredExporters, &exporters.JsonExporter{},
		)
	}

	if len(configuredExporters) == 0 {
		log.Warn().Msg("No exporter configured, using dummy exporter.")
		return &exporters.DummyExporter{}
	} else if len(configuredExporters) == 1 {
		return configuredExporters[0]
	} else {
		return &exporters.ChainExporter{
			Exporters:  configuredExporters,
			Marshalled: make([][]byte, len(configuredExporters)),
		}
	}
}

// CheckError wraps a function call which returns an error.
// If an error is returned, then `os.Exit(1)` is called
func CheckError(err error) {
	if err != nil {
		fmt.Printf("\nFatal error: %s\n", err)
		os.Exit(1)
	}
}

// RunBenchmark actually performs the task of running a benchmark.
func RunBenchmark(bench define.Benchmarkable) {
	var (
		cfg      *define.Config = define.GetConfig()
		exporter define.Exporterable
	)
	SetLogLevel(cfg)
	LogVersion()
	exporter = GetExporter(cfg)

	CheckError(exporter.Setup(cfg))
	CheckError(exporter.Healthcheck())
	CheckError(bench.Setup(cfg))
	CheckError(bench.Run(exporter))

	// Don't want to exit on these, as doing so
	// would interrupt other cleanup tasks
	bench.Teardown(cfg)
	exporter.Teardown()
}
