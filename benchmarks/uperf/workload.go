//go:build uperf
// +build uperf

// xml.go defines functionality for parsing XML workload files into structs.
// References:
// - https://tutorialedge.net/golang/parsing-xml-with-golang/
// - http://uperf.org/manual.html#id2547295
package uperf

import (
	"encoding/xml"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

type Profile struct {
	Name   string  `xml:"name,attr"`
	Groups []Group `xml:"group"`
}

type Group struct {
	NThreads     string        `xml:"nthreads,attr"`
	Transactions []Transaction `xml:"transaction"`
}

type Transaction struct {
	Duration   string   `xml:"duration,attr"`
	Iterations string   `xml:"iterations,attr"`
	FlowOps    []FlowOp `xml:"flowop"`
}

type FlowOp struct {
	Type    string `xml:"type,attr"`
	Options string `xml:"options,attr"`
}

type FlowOpOptions map[string]string

// ParseFlowOpOptions parses the option attr of a FlowOp.
// These options are presented as space-separated key=value pairs.
func ParseFlowOpOptions(optionsStr string) (FlowOpOptions, error) {
	options := FlowOpOptions{}
	var split []string

	for _, optionStr := range strings.Fields(optionsStr) {
		split = strings.Split(optionStr, "=")
		if len(split) != 2 {
			return nil, fmt.Errorf(
				"unable to parse option, expected 'key=value': %s", optionStr,
			)
		}
		options[split[0]] = split[1]
	}

	return options, nil
}

// PerformEnvSubst finds environment variables defined in the workload xml
// and tries to substitute them with their values within the environment.
func PerformEnvSubst(workloadRawBytes []byte) ([]byte, error) {
	workloadRawString := string(workloadRawBytes)

	r := regexp.MustCompile(`\$[a-zA-Z0-9]+`)
	matches := r.FindAllString(workloadRawString, -1)
	if matches == nil {
		return workloadRawBytes, nil
	}

	// Need to sort so longer matches get substituted first.
	// This prevents values such as `$h2` getting turned into `${h}2`
	sort.Slice(matches, func(i, j int) bool {
		return len(matches[i]) > len(matches[j])
	})

	for _, envVarInXML := range matches {
		envVar := strings.TrimSpace(envVarInXML)
		envVar = envVar[1:] // remove the first '$'
		envVal, ok := os.LookupEnv(envVar)
		if !ok {
			return workloadRawBytes, fmt.Errorf(
				"unable to find value for environment variable %s",
				envVar,
			)
		}
		workloadRawString = strings.ReplaceAll(workloadRawString, envVarInXML, envVal)
	}

	return []byte(workloadRawString), nil
}

// ParseWorkloadXML parses a uperf workload xml file into a Profile struct.
// Each workload xml file only contains a single Profile, per the uperf documentation.
// See http://uperf.org/manual.html#id2547203
func ParseWorkloadXML(workloadRawBytes []byte) (*Profile, error) {
	var profile Profile
	err := xml.Unmarshal(workloadRawBytes, &profile)
	if err != nil {
		return nil, err
	}
	return &profile, nil
}
