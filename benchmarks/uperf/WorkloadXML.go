// xml.go defines functionality for parsing XML workload files into structs.
// References:
// - https://tutorialedge.net/golang/parsing-xml-with-golang/
// - http://uperf.org/manual.html#id2547295
package uperf

import (
	"encoding/xml"
	"fmt"
	"strings"
)

type Profile struct {
	// XMLName xml.Name `xml:"profile"`
	Name   string  `xml:"name,attr"`
	Groups []Group `xml:"group"`
}

type Group struct {
	// XMLName      xml.Name      `xml:"group"`
	NThreads     string        `xml:"nthreads,attr"`
	Transactions []Transaction `xml:"transaction"`
}

type Transaction struct {
	// XMLName    xml.Name `xml:"transaction"`
	Duration   string   `xml:"duration,attr"`
	Iterations string   `xml:"iterations,attr"`
	FlowOps    []FlowOp `xml:"flowop"`
}

type FlowOp struct {
	// XMLName xml.Name `xml:"flowop"`
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
