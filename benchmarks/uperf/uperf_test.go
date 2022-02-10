//go:build uperf_test
// +build uperf_test

package uperf

import (
	"testing"

	"github.com/learnitall/gobench/define"
)

// TestUperfBenchmarkImplementsBenchmarkable ensures that the UperfBenchmark
// object implements the Benchmarkable interface
func TestUperfBenchmarkImplementsBenchmarkable(t *testing.T) {
	var uperf interface{} = &UperfBenchmark{}
	_, ok := uperf.(define.Benchmarkable)

	// Can use this line to help debug problems within IDE
	// var _ define.Benchmarkable = &UperfBenchmark{}

	if !ok {
		t.Errorf(
			"UperfBenchmark failed Benchmarkable type assertion",
		)
	}
}
