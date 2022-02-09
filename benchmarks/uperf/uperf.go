package uperf

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"

	"github.com/learnitall/gobench/define"
	"github.com/rs/zerolog/log"
)

// UperfResultPayload holds results of the uperf benchmark, ready to be marshalled and exported.
type UperfResultPayload struct {
	Result   UperfStdout
	Profile  Profile
	Cmd      []string
	Metadata define.Metadata
}

// UperfBenchmark helps facilitate running Uperf.
// It implements the define.Benchmarkable interface.
type UperfBenchmark struct {
	WorkloadPath string
	WorkloadRaw  string
	Profile      Profile
	Cmd          []string
	Metadata     define.Metadata
}

// Setup runs setup tasks for the UperfBenchmark.
// Assumes that the following fields are already set:
// - WorkloadPath
// - Cmd
func (u *UperfBenchmark) Setup(cfg *define.Config) error {
	workloadBytes, err := ioutil.ReadFile(u.WorkloadPath)
	if err != nil {
		return fmt.Errorf(
			"unable to read workload file at %s: %s",
			u.WorkloadPath, err,
		)
	}

	profile, err := ParseWorkloadXML(workloadBytes)
	if err != nil {
		return fmt.Errorf(
			"unable to parse workload file at %s: %s",
			u.WorkloadPath, err,
		)
	}

	u.WorkloadRaw = string(workloadBytes)
	u.Profile = *profile
	u.Metadata = define.GetMetadataPayload(cfg)

	return nil
}

// Run facilitates running, parsing and exporting data from the uperf benchmark.
// Assumes Setup has already been called prior.
func (u *UperfBenchmark) Run(exporter define.Exporterable) error {
	log.Info().
		Str("cmd", strings.Join(u.Cmd, " ")).
		Str("workload_path", u.WorkloadPath).
		Msg("Running Uperf")

	cmd := exec.Command(u.Cmd[0], u.Cmd[1:]...)
	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	stdout := out.String()
	if err != nil {
		log.Fatal().
			Str("stdout", stdout).
			Err(err).
			Msg("Error occurred while running uperf.")
		return err
	}

	log.Info().
		Msg("Uperf successfully finished, preparing results.")
	log.Debug().
		Str("stdout", stdout).
		Msg("Received the following stdout.")

	stdoutResult, err := ParseUperfStdout(stdout)
	if err != nil {
		log.Fatal().
			Str("stdout", stdout).
			Err(err).
			Msg("Received error while parsing uperf stdout.")
		return err
	}

	payload := UperfResultPayload{
		Result:   *stdoutResult,
		Profile:  u.Profile,
		Cmd:      u.Cmd,
		Metadata: u.Metadata,
	}
	log.Info().Msg("Parsed stdout and prepared payload, marshalling.")

	marshalled, err := exporter.Marshal(payload)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Unable to marshal uperf result payload.")
		return err
	}

	err = exporter.Export(marshalled)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Unexpected error while exporting marshalled payload.")
	}

	return nil
}

// Teardown function for the uperf benchmark.
// No specific tasks need to be run, so this just returns nil.
func (u *UperfBenchmark) Teardown(*define.Config) error {
	return nil
}
