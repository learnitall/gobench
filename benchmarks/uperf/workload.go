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
	"strconv"
	"strings"
	"time"
)

type profileXML struct {
	Name   string     `xml:"name,attr"`
	Groups []groupXML `xml:"group"`
}

type groupXML struct {
	NThreads     string           `xml:"nthreads,attr"`
	Transactions []transactionXML `xml:"transaction"`
}

type transactionXML struct {
	Duration   string      `xml:"duration,attr"`
	Iterations string      `xml:"iterations,attr"`
	FlowOps    []flowOpXML `xml:"flowop"`
}

type flowOpXML struct {
	Type    string `xml:"type,attr"`
	Options string `xml:"options,attr"`
}

type Profile struct {
	Name   string
	Groups []Group
}

type Group struct {
	NThreads     int
	Transactions []Transaction
}

type Transaction struct {
	DurationSeconds int
	Iterations      int
	FlowOps         []flowOpXML
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
	var (
		err                error
		profile            Profile = Profile{}
		profileXMLInstance profileXML
		nthreads           int
		duration           int
		iterations         int
	)

	err = xml.Unmarshal(workloadRawBytes, &profileXMLInstance)
	if err != nil {
		return nil, err
	}

	// Now we go through and parse each field into the Profile
	// I know this for loop is mad disgusting, but I'm not sure how
	// else to do this differently.
	profile.Name = profileXMLInstance.Name
	profile.Groups = []Group{}

	for _, groupXML := range profileXMLInstance.Groups {
		group := Group{}
		nthreads, err = strconv.Atoi(groupXML.NThreads)
		if err != nil {
			return nil, fmt.Errorf(
				"unable to parse nthreads string into an int: %s",
				groupXML.NThreads,
			)
		}
		group.NThreads = nthreads
		group.Transactions = []Transaction{}
		for _, transactionXML := range groupXML.Transactions {
			duration = -1
			iterations = -1
			if len(transactionXML.Duration) > 0 {
				_duration, err := time.ParseDuration(transactionXML.Duration)
				if err != nil {
					return nil, fmt.Errorf(
						"Unable to parse transaction duration '%s': %s",
						transactionXML.Duration,
						err,
					)
				}
				duration = int(_duration.Seconds())
			}
			if err == nil && len(transactionXML.Iterations) > 0 {
				iterations, err = strconv.Atoi(transactionXML.Iterations)
				if err != nil {
					return nil, fmt.Errorf(
						"Unable to parse transaction iterations '%s': %s",
						transactionXML.Iterations,
						err,
					)
				}
			}
			group.Transactions = append(
				group.Transactions,
				Transaction{
					DurationSeconds: duration,
					Iterations:      iterations,
					FlowOps:         transactionXML.FlowOps,
				},
			)
		}
		profile.Groups = append(profile.Groups, group)
	}

	return &profile, nil
}
