package uperf

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/docker/go-units"
)

type DetailsStat struct {
	Name           string
	TotalBytes     int64
	TotalSeconds   float64
	BytesPerSecond int64
	OpsPerSecond   int
}

type DetailsStatRaw struct {
	TimestampMS float64
	Name        string
	Bytes       int
	Ops         int
}

type GroupDetails struct {
}

type TXStats struct {
}

type FlowopStats struct {
}

type NetstatStats struct {
}

type RunStats struct {
}

type UperfStdout struct {
	DetailsStats      []DetailsStat
	TXCommandStatsRaw []DetailsStatRaw
}

// getParseError constructs a new error instance for when a struct's
// field cannot be parser correctly.
func getParseError(stdoutLine string, field string, object string, name string) error {
	return fmt.Errorf(
		"unable to parse %s from %s, line: %s, field: %s",
		name,
		object,
		stdoutLine,
		field,
	)
}

// ParseDetailsStat parses the given line expected to contain a DetailsStat.
// Checks that the line starts with 'Txn' and contains 7 fields.
// Examples:
// - Txn1            0 /   1.00(s) =            0           1op/s
// - Txn2      67.14GB /  30.23(s) =    19.08Gb/s      291075op/s
// - Txn3            0 /   0.00(s) =            0           0op/s
func ParseDetailsStat(stdoutLine string) (*DetailsStat, error) {
	fields := strings.Fields(stdoutLine)
	if len(fields) != 7 || !strings.HasPrefix(stdoutLine, "Txn") {
		return nil, fmt.Errorf(
			"expected DetailsStat line to start with 'Txn' and contain 7 fields: %s",
			stdoutLine,
		)
	}

	totalBytes, err := units.FromHumanSize(fields[1])
	if err != nil {
		return nil,
			getParseError(stdoutLine, fields[1], "DetailsStat", "TotalBytes")
	}

	totalTime, err := time.ParseDuration(fields[3])
	if err != nil {
		return nil,
			getParseError(stdoutLine, fields[3], "DetailsStat", "TotalSeconds")
	}

	bytesPerSecond, err := units.FromHumanSize(
		strings.Replace(fields[5], "/s", "", 1),
	)
	if err != nil {
		return nil,
			getParseError(stdoutLine, fields[5], "DetailsStat", "BytesPerSecond")
	}

	opsPerSecond, err := strconv.Atoi(strings.Replace(fields[6], "op/s", "", 1))
	if err != nil {
		return nil,
			getParseError(stdoutLine, fields[6], "DetailsStat", "OpsPerSecond")
	}

	return &DetailsStat{
		Name:           fields[0],
		TotalBytes:     totalBytes,
		TotalSeconds:   totalTime.Seconds(),
		BytesPerSecond: bytesPerSecond,
		OpsPerSecond:   opsPerSecond,
	}, nil
}

// ParseDetailsStatRaw parses the given line expected to contain a DetailsStatRaw.
// Checks that the line starts with `timestamp_ms` and has four fields.
func ParseDetailsStatRaw(stdoutLine string) (*DetailsStatRaw, error) {
	fields := strings.Fields(stdoutLine)
	if len(fields) != 4 || !strings.HasPrefix(stdoutLine, "timestamp_ms") {
		return nil, fmt.Errorf(
			"expected DetailsStatRaw line to start with 'timestamp_ms' and contain 4 fields: %s",
			stdoutLine,
		)
	}

	timestamp_ms_fields := strings.Split(fields[0], ":")
	if len(timestamp_ms_fields) != 2 {
		return nil,
			getParseError(stdoutLine, fields[0], "DetailsStatRaw", "TimestampMS")
	}
	timestamp_ms, err := strconv.ParseFloat(timestamp_ms_fields[1], 64)
	if err != nil {
		return nil,
			getParseError(stdoutLine, fields[0], "DetailsStatRaw", "TimestampMS")
	}

	name_fields := strings.Split(fields[1], ":")
	if len(name_fields) != 2 {
		return nil,
			getParseError(stdoutLine, fields[1], "DetailsStatRaw", "Name")
	}
	name := name_fields[1]

	totalByteFields := strings.Split(fields[2], ":")
	if len(totalByteFields) != 2 {
		return nil,
			getParseError(stdoutLine, fields[2], "DetailsStatRaw", "Bytes")
	}
	totalBytes, err := strconv.Atoi(totalByteFields[1])
	if err != nil {
		return nil,
			getParseError(stdoutLine, fields[2], "DetailsStatRaw", "Bytes")
	}

	totalOpsFields := strings.Split(fields[3], ":")
	if len(totalOpsFields) != 2 {
		return nil,
			getParseError(stdoutLine, fields[3], "DetailsStatRaw", "Ops")
	}
	totalOps, err := strconv.Atoi(totalOpsFields[1])
	if err != nil {
		return nil,
			getParseError(stdoutLine, fields[3], "DetailsStatRaw", "Ops")
	}

	return &DetailsStatRaw{
		Name:        name,
		Bytes:       totalBytes,
		Ops:         totalOps,
		TimestampMS: timestamp_ms,
	}, nil
}
