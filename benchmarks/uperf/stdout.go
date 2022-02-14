//go:build uperf
// +build uperf

package uperf

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/docker/go-units"
	"github.com/learnitall/gobench/define"
	"github.com/rs/zerolog/log"
)

// StatSectionType defines the section a stat was parsed from.
type StatSectionType string

const (
	StatSectionGroup     StatSectionType = "groups"
	StatSectionStrand    StatSectionType = "strand"
	StatSectionTX        StatSectionType = "tx"
	StatSectionFlowopAvg StatSectionType = "flowop_avg"
	StatSectionTxAvg     StatSectionType = "tx_avg"
	StatSectionNetstat   StatSectionType = "netstat"
	StatSectionRun       StatSectionType = "run"
	StatSectionRunDiff   StatSectionType = "run_diff"
)

// DetailsFormat tracks whether the given stat was parsed using uperfs 'raw' format (ie when -R is given) or 'computed' format
type DetailsFormat string

const (
	DetailsFormatRaw      DetailsFormat = "raw"
	DetailsFormatComputed DetailsFormat = "computed"
)

type DetailsStat struct {
	Name         string
	Metadata     *define.Metadata
	SectionType  StatSectionType
	DetailFormat DetailsFormat
	// Computed Stats
	TotalBytes     int64   `json:",omitempty"`
	TotalSeconds   float64 `json:",omitempty"`
	BytesPerSecond int64   `json:",omitempty"`
	OpsPerSecond   int     `json:",omitempty"`
	// Raw Stats
	TimestampMS float64 `json:",omitempty"`
	Bytes       int     `json:",omitempty"`
	Ops         int     `json:",omitempty"`
}

type AveragesStat struct {
	Name        string
	Metadata    *define.Metadata
	SectionType StatSectionType
	Count       int64
	AvgSeconds  float64
	CpuSeconds  float64
	MaxSeconds  float64
	MinSeconds  float64
}

type NetstatStat struct {
	Name              string
	Metadata          *define.Metadata
	SectionType       StatSectionType
	OutPktsPerSecond  int64
	InPktsPerSecond   int64
	OutBytesPerSecond int64
	InBytesPerSecond  int64
}

type RunStat struct {
	Hostname                 string
	Metadata                 *define.Metadata
	SectionType              StatSectionType
	TimeSeconds              float64
	DataBytes                int64
	ThroughputBytesPerSecond int64
	Operations               int64
	Errors                   float64
}

type RunDiffStat struct {
	Metadata                  *define.Metadata
	SectionType               StatSectionType
	TimeDeltaPercentage       float64
	DataDeltaPercentage       float64
	ThroughputDeltaPercentage float64
	OperationsDeltaPercentage float64
	ErrorsDeltaPercentage     float64
}

// UperfStdout represents the data parsed from uperf's stdout.
// Acts as a list of pointers to any struct that should be marshalled and exported.
type UperfStdout []interface{}

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

// parseSecondsString parses a string of the format <float>(s) or <float>s, returning a corresponding float64.
func parseSecondsString(humanString string) (float64, error) {
	duration, err := time.ParseDuration(
		strings.Replace(
			strings.Replace(humanString, "(", "", 1),
			")", "", 1,
		),
	)

	if err != nil {
		return 0, err
	}
	return duration.Seconds(), nil
}

// parseDataString parses a string of the format <float><data units>, normalizing it to bytes.
func parseDataString(humanString string) (int64, error) {
	return units.FromHumanSize(humanString)
}

// parseDataPerSecond parses a string of the format '<float><data units>/s', normalizing
// it to bytes per second.
func parseDataPerSecondString(humanString string) (int64, error) {
	return parseDataString(
		strings.Replace(humanString, "/s", "", 1),
	)
}

// checkNumFields checks if the number of fields in the given string is greater than or equal to a given number.
// If the number of fields is not greater than or equal to the expected amount then an error
// is returned with the following message:
// `expected <object> line to contain at least <expected> fields, instead given line has <num fields in humanString>`.
// We check for the number of fields to be greater than or equal to the given
// value because Uperf does not consistently put debug output on its own line.
func checkNumFields(humanString string, expected int, object string) ([]string, error) {
	fields := strings.Fields(humanString)
	if len(fields) < expected {
		return nil, fmt.Errorf(
			"expected %s line to contain at least %d fields, instead given line has %d: %s",
			object, expected, len(fields), humanString,
		)
	}
	return fields, nil
}

// checkPrefix checks if the prefix of the given string is equal to a given substring.
// If the given string does not have a prefix equal to the given substring,
// than an error is returned with the following message:
// `expected <object> line to start with <expected>: <humanString>`
func checkPrefix(humanString string, expected string, object string) error {
	if !strings.HasPrefix(humanString, expected) {
		return fmt.Errorf(
			"expected %s line to start with %s: %s",
			object, expected, humanString,
		)
	}
	return nil
}

// addMetadataField adds the given pointer to a define.Metadata object into the
// given struct pointer, if the struct has a field named 'Metadata'.
// If the Metadata field exists on the given struct and it can be added, then
// a debug log is emitted.
func addMetadataField(targetStruct interface{}, metadata define.Metadata) {
	metadataField := reflect.ValueOf(targetStruct).Elem().FieldByName("Metadata")
	if metadataField.CanSet() {
		metadataField.Set(reflect.ValueOf(&metadata))
		log.Debug().
			Interface("stat", targetStruct).
			Msg("Added metadata to stdout stat.")
	}
}

// ParseDetailsStatComputed parses the given line expected to contain a DetailsStat in a computed format.
// Checks that the line contains 7 fields.
// Examples:
// - `Txn1            0 /   1.00(s) =            0           1op/s`
// - `Txn2      67.14GB /  30.23(s) =    19.08Gb/s      291075op/s`
// - `Txn3            0 /   0.00(s) =            0           0op/s`
func parseDetailsStatComputed(stdoutLine string) (*DetailsStat, error) {
	fields, err := checkNumFields(stdoutLine, 7, "DetailsStatComputed")
	if err != nil {
		return nil, err
	}

	_onError := func(fieldNum int, name string) (*DetailsStat, error) {
		return nil,
			getParseError(stdoutLine, fields[fieldNum], "DetailsStatComputed", name)
	}

	totalBytes, err := units.FromHumanSize(fields[1])
	if err != nil {
		return _onError(1, "TotalBytes")
	}

	totalTime, err := parseSecondsString(fields[3])
	if err != nil {
		return _onError(3, "TotalSeconds")
	}

	bytesPerSecond, err := parseDataPerSecondString(fields[5])
	if err != nil {
		return _onError(5, "BytesPerSecond")
	}

	opsPerSecond, err := strconv.Atoi(strings.Replace(fields[6], "op/s", "", 1))
	if err != nil {
		return _onError(6, "OpsPerSecond")
	}

	return &DetailsStat{
		Name:           fields[0],
		DetailFormat:   DetailsFormatComputed,
		TotalBytes:     totalBytes,
		TotalSeconds:   totalTime,
		BytesPerSecond: bytesPerSecond,
		OpsPerSecond:   opsPerSecond,
	}, nil
}

// ParseDetailsStatRaw parses the given line expected to contain a DetailsStat in a raw format.
// Checks that the line starts with `timestamp_ms` and has four fields.
// Examples:
// - `timestamp_ms:1644254626628.9016 name:Group0 nr_bytes:7995523072 nr_ops:9760162`
// - `timestamp_ms:1644254626628.8372 name:Txn2 nr_bytes:79955230720 nr_ops:9760160`
func parseDetailsStatRaw(stdoutLine string) (*DetailsStat, error) {
	err := checkPrefix(stdoutLine, "timestamp_ms", "DetailsStatRaw")
	if err != nil {
		return nil, err
	}
	fields, err := checkNumFields(stdoutLine, 4, "DetailsStatRaw")
	if err != nil {
		return nil, err
	}

	_onError := func(fieldNum int, name string) (*DetailsStat, error) {
		return nil,
			getParseError(stdoutLine, fields[fieldNum], "DetailsStatRaw", name)
	}

	timestamp_ms_fields := strings.Split(fields[0], ":")
	if len(timestamp_ms_fields) != 2 {
		return _onError(0, "TimestampMS")
	}
	timestamp_ms, err := strconv.ParseFloat(timestamp_ms_fields[1], 64)
	if err != nil {
		return _onError(0, "TimestampMS")
	}

	name_fields := strings.Split(fields[1], ":")
	if len(name_fields) != 2 {
		return _onError(1, "Name")
	}
	name := name_fields[1]

	totalByteFields := strings.Split(fields[2], ":")
	if len(totalByteFields) != 2 {
		return _onError(2, "Bytes")
	}
	totalBytes, err := strconv.Atoi(totalByteFields[1])
	if err != nil {
		return _onError(2, "Bytes")
	}

	totalOpsFields := strings.Split(fields[3], ":")
	if len(totalOpsFields) != 2 {
		return _onError(3, "Ops")
	}
	totalOps, err := strconv.Atoi(totalOpsFields[1])
	if err != nil {
		return _onError(3, "Ops")
	}

	return &DetailsStat{
		Name:         name,
		DetailFormat: DetailsFormatRaw,
		Bytes:        totalBytes,
		Ops:          totalOps,
		TimestampMS:  timestamp_ms,
	}, nil
}

// ParseDetailsStat is an alias for ParseDetailsStatRaw and ParseDetailsStatComputed.
// If the given line starts with "timestamp_ms", then ParseDetailsStatRaw is called, otherwise ParseDetailsStatComputed will be called.
func parseDetailsStat(stdoutLine string) (*DetailsStat, error) {
	if strings.HasPrefix(stdoutLine, "timestamp_ms") {
		return parseDetailsStatRaw(stdoutLine)
	}
	return parseDetailsStatComputed(stdoutLine)
}

// ParseAveragesStat parses the given line expected to contain an AveragesStat.
// Checks that the given line has six fields.
// Examples:
// - `connect                1     69.62us      0.00ns     69.62us     69.62us`
// - `Txn2                   1      1.30us      0.00ns      1.30us      1.30us`
func parseAveragesStat(stdoutLine string) (*AveragesStat, error) {
	fields, err := checkNumFields(stdoutLine, 6, "AveragesStat")
	if err != nil {
		return nil, err
	}

	_onError := func(fieldNum int, name string) (*AveragesStat, error) {
		return nil,
			getParseError(stdoutLine, fields[fieldNum], "AveragesStat", name)
	}

	name := fields[0]

	count, err := strconv.Atoi(fields[1])
	if err != nil {
		return _onError(1, "Count")
	}

	avg, err := parseSecondsString(fields[2])
	if err != nil {
		return _onError(2, "AvgSeconds")
	}

	cpu, err := parseSecondsString(fields[3])
	if err != nil {
		return _onError(3, "CpuSeconds")
	}

	max, err := parseSecondsString(fields[4])
	if err != nil {
		return _onError(4, "MaxSeconds")
	}

	min, err := parseSecondsString(fields[5])
	if err != nil {
		return _onError(5, "MinSeconds")
	}

	return &AveragesStat{
		Name:       name,
		Count:      int64(count),
		AvgSeconds: avg,
		CpuSeconds: cpu,
		MaxSeconds: max,
		MinSeconds: min,
	}, nil
}

// ParseNetstatStat parses the given line expected to contain a NetstatStat.
// Checks that the line has five fields.
// Examples:
// - `lo         307537      307537    20.22Gb/s    20.22Gb/s`
// - `lo         272158      272158    17.90Gb/s    17.90Gb/s`
func parseNetstatStat(stdoutLine string) (*NetstatStat, error) {
	fields, err := checkNumFields(stdoutLine, 5, "NetstatStat")
	if err != nil {
		return nil, err
	}

	_onError := func(fieldNum int, name string) (*NetstatStat, error) {
		return nil, getParseError(stdoutLine, fields[fieldNum], "NetstatStat", name)
	}

	name := fields[0]

	outPkts, err := strconv.Atoi(fields[1])
	if err != nil {
		return _onError(1, "OutPktsPerSecond")
	}

	inPkts, err := strconv.Atoi(fields[2])
	if err != nil {
		return _onError(2, "InPktsPerSecond")
	}

	outBytes, err := parseDataPerSecondString(fields[3])
	if err != nil {
		return _onError(3, "OutBytesPerSecond")
	}

	inBytes, err := parseDataPerSecondString(fields[4])
	if err != nil {
		return _onError(4, "InBytesPerSecond")
	}

	return &NetstatStat{
		Name:              name,
		OutPktsPerSecond:  int64(outPkts),
		InPktsPerSecond:   int64(inPkts),
		OutBytesPerSecond: outBytes,
		InBytesPerSecond:  inBytes,
	}, nil
}

// ParseRunStat parses the given line expected to contain a RunStat.
// Checks that the line contains six fields.
// Examples:
// - `127.0.0.1         32.34s    59.48GB    15.80Gb/s      7795963        0.00`
// - `master            32.34s    67.14GB    17.84Gb/s      8800383        0.00`
func parseRunStat(stdoutLine string) (*RunStat, error) {
	fields, err := checkNumFields(stdoutLine, 6, "RunStat")
	if err != nil {
		return nil, err
	}

	_onError := func(field int, name string) (*RunStat, error) {
		return nil, getParseError(stdoutLine, fields[field], "RunStat", name)
	}

	hostname := fields[0]

	timeSeconds, err := parseSecondsString(fields[1])
	if err != nil {
		return _onError(1, "TimeSeconds")
	}

	dataBytes, err := parseDataString(fields[2])
	if err != nil {
		return _onError(2, "DataBytes")
	}

	throughput, err := parseDataPerSecondString(fields[3])
	if err != nil {
		return _onError(3, "ThroughputBytesPerSecond")
	}

	ops, err := strconv.Atoi(fields[4])
	if err != nil {
		return _onError(4, "Operations")
	}

	es, err := strconv.ParseFloat(fields[5], 64)
	if err != nil {
		return _onError(5, "Errors")
	}

	return &RunStat{
		Hostname:                 hostname,
		TimeSeconds:              timeSeconds,
		DataBytes:                dataBytes,
		ThroughputBytesPerSecond: throughput,
		Operations:               int64(ops),
		Errors:                   es,
	}, nil
}

// ParseRunDiffStat parses the given line expected to contain a RunDiffStat.
// Checks the line contains six fields and starts with "Differenc(%)"
// Examples:
// - `Difference(%)     -0.00%     11.41%       11.41%       11.41%       0.00%`
// - `Difference(%)     -0.00%     11.66%       11.66%       11.66%       0.00%`
func parseRunDiffStat(stdoutLine string) (*RunDiffStat, error) {
	err := checkPrefix(stdoutLine, "Difference(%)", "RunDiffStat")
	if err != nil {
		return nil, err
	}

	// Get rid of these so we can parse our float values
	// more easily
	stdoutLine = strings.ReplaceAll(stdoutLine, "%", "")

	fields, err := checkNumFields(stdoutLine, 6, "RunDiffStat")
	if err != nil {
		return nil, err
	}

	_onError := func(fieldNum int, name string) (*RunDiffStat, error) {
		return nil,
			getParseError(stdoutLine, fields[fieldNum], "RunDiffStat", name)
	}

	// These are all floats and can be parsed the same way
	fieldNames := []string{
		"TimeDeltaPercentage",
		"DataDeltaPercentage",
		"ThroughputDeltaPercentage",
		"OperationsDeltaPercentage",
		"ErrorsDeltaPercentage",
	}
	RunDiffStat := RunDiffStat{}
	var fieldIndex int

	for i, fieldName := range fieldNames {
		fieldIndex = i + 1
		value, err := strconv.ParseFloat(fields[fieldIndex], 64)
		if err != nil {
			return _onError(fieldIndex, fieldName)
		}
		reflect.
			ValueOf(&RunDiffStat).
			Elem().
			FieldByName(fieldName).
			SetFloat(value)
	}

	return &RunDiffStat, nil
}

// parseUperfStdoutSection calls the given callback function on each line
// in the given stdoutLines until a line that is empty is hit.
// If a line is encountered that contains dashes, it will be skipped over.
// We assume that the stdoutLines starts with the section header/title, ie:
// `[Section Header, -(a bunch of dashes)-, Section Contents]`
func parseUperfStdoutSection(stdoutLines []string, callback func(string) error) ([]string, error) {
	// Skip ahead two lines to get past the header and dashes
	stdoutLines = stdoutLines[1:]
	for {
		stdoutLines = stdoutLines[1:]

		if len(stdoutLines) == 0 {
			return stdoutLines, nil
		}

		if len(stdoutLines[0]) == 0 {
			return stdoutLines, nil
		}

		if strings.HasPrefix(stdoutLines[0], "---") {
			continue
		}

		err := callback(stdoutLines[0])
		if err != nil {
			return stdoutLines, err
		}
	}
}

// parseUperfStdoutDetailsSection parses a section containing DetailsStat.
func parseUperfStdoutDetailsSection(stdoutLines []string, sectionType StatSectionType, result *UperfStdout) ([]string, error) {
	return parseUperfStdoutSection(
		stdoutLines,
		func(nextLine string) error {
			detailsStat, err := parseDetailsStat(nextLine)
			if err != nil {
				return err
			}
			detailsStat.SectionType = sectionType
			*result = append(*result, detailsStat)
			return nil
		},
	)
}

// parseUperfStdoutAveragesSection parses a section containing AveragesStat.
func parseUperfStdoutAveragesSection(stdoutLines []string, sectionType StatSectionType, result *UperfStdout) ([]string, error) {
	return parseUperfStdoutSection(
		stdoutLines,
		func(nextLine string) error {
			averagesStat, err := parseAveragesStat(nextLine)
			if err != nil {
				return err
			}
			averagesStat.SectionType = sectionType
			*result = append(*result, averagesStat)
			return nil
		},
	)
}

// parseUperfStdoutNetstatSection parses a section containing NetstatStat,
// placing the results into the given UperfStdout struct under the `NetstatStats` map.
func parseUperfStdoutNetstatSection(stdoutLines []string, result *UperfStdout) ([]string, error) {
	return parseUperfStdoutSection(
		stdoutLines,
		func(nextLine string) error {
			if strings.HasPrefix(strings.ReplaceAll(nextLine, " ", ""), "Nicopkts/s") {
				return nil
			}
			netstatStat, err := parseNetstatStat(nextLine)
			if err != nil {
				return err
			}
			netstatStat.SectionType = StatSectionNetstat
			*result = append(*result, netstatStat)
			return nil
		},
	)
}

// parseUperfStdoutRunStatsSection parses a section containing RunStats,
// placing the results into the given UperfStdout.
func parseUperfStdoutRunStatsSection(stdoutLines []string, result *UperfStdout) ([]string, error) {
	return parseUperfStdoutSection(
		stdoutLines,
		func(nextLine string) error {
			if strings.HasPrefix(strings.ReplaceAll(nextLine, " ", ""), "HostnameTime") {
				return nil
			}
			if strings.HasPrefix(nextLine, "Difference") {
				runStatsDiff, err := parseRunDiffStat(nextLine)
				if err != nil {
					return err
				}
				runStatsDiff.SectionType = StatSectionRunDiff
				*result = append(*result, runStatsDiff)
				return nil
			}
			runStat, err := parseRunStat(nextLine)
			if err != nil {
				return err
			}
			runStat.SectionType = StatSectionRun
			*result = append(*result, runStat)
			return nil
		},
	)
}

func parseUperfStdout(stdoutLines []string, result *UperfStdout) error {
	if len(stdoutLines) == 0 {
		return nil
	}

	currentLine := stdoutLines[0]

	var (
		detailsStat *DetailsStat = nil
		err         error        = nil
	)

	if strings.HasPrefix(currentLine, "Group Details") {
		log.Debug().Str("current_line", currentLine).Msg("Parsing Group Details section")
		stdoutLines, err = parseUperfStdoutDetailsSection(
			stdoutLines, StatSectionGroup, result,
		)
		if err != nil {
			return err
		}
	} else if strings.HasPrefix(currentLine, "Strand Details") {
		log.Debug().Str("current_line", currentLine).Msg("Parsing Strand Details section")
		stdoutLines, err = parseUperfStdoutDetailsSection(
			stdoutLines, StatSectionStrand, result,
		)
		if err != nil {
			return err
		}
	} else if strings.HasPrefix(strings.ReplaceAll(currentLine, " ", ""), "TxnCount") {
		log.Debug().Str("current_line", currentLine).Msg("Parsing Txn Averages section")
		stdoutLines, err = parseUperfStdoutAveragesSection(
			stdoutLines, StatSectionTxAvg, result,
		)
		if err != nil {
			return err
		}
	} else if strings.HasPrefix(currentLine, "Flowop") {
		log.Debug().Str("current_line", currentLine).Msg("Parsing Flowop Averages section")
		stdoutLines, err = parseUperfStdoutAveragesSection(
			stdoutLines, StatSectionFlowopAvg, result,
		)
		if err != nil {
			return err
		}
	} else if strings.HasPrefix(currentLine, "Netstat statistics") {
		log.Debug().Str("current_line", currentLine).Msg("Parsing Netstat Statistics section")
		stdoutLines, err = parseUperfStdoutNetstatSection(
			stdoutLines, result,
		)
		if err != nil {
			return err
		}
	} else if strings.HasPrefix(currentLine, "Run Statistics") {
		log.Debug().Str("current_line", currentLine).Msg("Parsing Run Statistics section")
		stdoutLines, err = parseUperfStdoutRunStatsSection(
			stdoutLines, result,
		)
		if err != nil {
			return err
		}
	} else if strings.HasPrefix(currentLine, "Txn") || strings.HasPrefix(currentLine, "Total") {
		log.Debug().Str("current_line", currentLine).Msg("Parsing Txn detail (computed)")
		detailsStat, err = parseDetailsStatComputed(currentLine)
		detailsStat.SectionType = StatSectionTX
		if err != nil {
			return err
		}
		*result = append(*result, detailsStat)
		stdoutLines = stdoutLines[1:]
	} else if strings.HasPrefix(currentLine, "timestamp_ms") {
		log.Debug().Str("current_line", currentLine).Msg("Parsing Txn detail (raw)")
		detailsStat, err = parseDetailsStatRaw(currentLine)
		detailsStat.SectionType = StatSectionTX
		if err != nil {
			return err
		}
		*result = append(*result, detailsStat)
		stdoutLines = stdoutLines[1:]
	} else {
		log.Debug().Str("current_line", currentLine).Msg("Skipping line")
		stdoutLines = stdoutLines[1:]
	}

	return parseUperfStdout(stdoutLines, result)
}

func ParseUperfStdout(uperfStdout string) (*UperfStdout, error) {
	var (
		lines  []string     = strings.Split(uperfStdout, "\n")
		result *UperfStdout = &UperfStdout{}
	)
	err := parseUperfStdout(lines, result)
	if err != nil {
		return nil, err
	}
	return result, err
}
