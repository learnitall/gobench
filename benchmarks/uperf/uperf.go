package uperf

import (
	"github.com/learnitall/gobench/define"
)

// UperfBenchmark helps facilitate running Uperf.
// It implements the define.Benchmarkable interface.
type UperfBenchmark struct {
	WorkloadPath string
}

func (u *UperfBenchmark) Run(*define.Config) error {
	return nil
}
