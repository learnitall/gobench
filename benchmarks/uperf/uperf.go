package uperf

import (
	"github.com/learnitall/gobench/define"
)

// UperfResultPayload holds results of the uperf benchmark, ready to be marshalled and exported.
type UperfResultPayload struct {
	OutputRaw  string
	ProfileRaw string
	Profile    Profile
	Cmd        string
}

// UperfBenchmark helps facilitate running Uperf.
// It implements the define.Benchmarkable interface.
type UperfBenchmark struct {
	WorkloadPath string
}

func (u *UperfBenchmark) Run(*define.Config) error {
	return nil
}
