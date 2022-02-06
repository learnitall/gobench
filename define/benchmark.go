// benchmark.go defines items relevent to the execution of benchmarks and the
// parsing of their output
package define

// Benchmarkable defines methods gobench needs to run a benchmark.
type Benchmarkable interface {
	// Run actually runs the benchmark and sends the benchmark's result
	// data to the Exporter. Assume benchmark passes if error is nil.
	Run(*Config) error
}
