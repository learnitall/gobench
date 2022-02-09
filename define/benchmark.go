// benchmark.go defines items relevent to the execution of benchmarks and the
// parsing of their output
package define

// Benchmarkable defines methods gobench needs to run a benchmark.
type Benchmarkable interface {
	// Setup runs tasks which need to be done prior to calling Run.
	// These tasks could involve verifying integrity of config files or
	// creating needed objects.
	Setup(*Config) error
	// Run actually runs the benchmark and sends the benchmark's result
	// data to the Exporter. Assume benchmark passes if error is nil.
	Run(Exporterable) error

	// Teardown does any finishing tasks after the benchmark has ran and
	// exported its results.
	Teardown(*Config) error
}
