package uperf

import (
	"encoding/json"
	"reflect"
	"testing"
)

// https://github.com/uperf/uperf/blob/7ac3d0b0353c42ea5c9c83c27f84b1873be50247/workloads/VoIPRx.xml
var WORKLOAD_XML_VOIPRX string = `<?xml version="1.0"?>
<profile name="VoIP Rx">
  <group nthreads="200">
      <transaction iterations="1">
        <flowop type="connect" options="remotehost=$h protocol=udp"/>
      </transaction>
      <transaction duration="120s">
            <flowop type="read" options="size=64"/>
      </transaction>
      <transaction iterations="1">
            <flowop type="disconnect" />
      </transaction>
  </group>

  <group nthreads="200">
      <transaction iterations="1">
        <flowop type="connect" options="remotehost=$h2 protocol=udp"/>
      </transaction>
      <transaction duration="120s">
            <flowop type="read" options="size=64"/>
      </transaction>
      <transaction iterations="1">
            <flowop type="disconnect" />
      </transaction>
  </group>
</profile>`
var WORKLOAD_JSON_VOIPRX = `{
	"name": "VoIP Rx",
	"groups": [
		{
			"nthreads": "200",
			"XMLName": "group",
			"transactions": [
				{
					"iterations": "1",
					"flowops": [
						{
						    "type": "connect",
							"options": "remotehost=$h protocol=udp"
						}
					]
				},
				{
					"duration": "120s",
					"flowops": [
						{
							"type": "read",
							"options": "size=64"
						}
					]
				},
				{
					"iterations": "1",
					"flowops": [
						{
							"type": "disconnect"
						}
					]
				}
			]
		},
		{
			"nthreads": "200",
			"transactions": [
				{
					"iterations": "1",
					"flowops": [
						{
							"type": "connect",
							"options": "remotehost=$h2 protocol=udp"
						}
					]
				},
				{
					"duration": "120s",
					"flowops": [
						{
							"type": "read",
							"options": "size=64"
						}
					]
				},
				{
					"iterations": "1",
					"flowops": [
						{
							"type": "disconnect"
						}
					]
				}
			]
		}
	]
}`

// https://raw.githubusercontent.com/uperf/uperf/7ac3d0b0353c42ea5c9c83c27f84b1873be50247/workloads/iperf.xml
var WORKLOAD_XML_IPERF string = `<?xml version="1.0"?>
<profile name="iPERF">
  <group nthreads="$nthr">
        <transaction iterations="1">
            <flowop type="connect" options="remotehost=$h protocol=$proto wndsz=50k  tcp_nodelay"/>
        </transaction>
        <transaction duration="30s">
            <flowop type="write" options="count=10 size=8k"/>
        </transaction>
        <transaction iterations="1">
            <flowop type="disconnect" />
        </transaction>
  </group>

</profile>`
var WORKLOAD_JSON_IPERF string = `{
	"name": "iPERF",
	"groups": [
		{
			"nthreads": "$nthr",
			"transactions": [
				{
					"iterations": "1",
					"flowops": [
						{
							"type": "connect",
							"options": "remotehost=$h protocol=$proto wndsz=50k  tcp_nodelay"
						}
					]
				},
				{
					"duration": "30s",
					"flowops": [
						{
							"type": "write",
							"options": "count=10 size=8k"
						}
					]
				},
				{
					"iterations": "1",
					"flowops": [
						{
							"type": "disconnect"
						}
					]
				}
			]
		}
	]
}`
var WORKLOAD_XML_TO_JSON_TESTS map[string]string = map[string]string{
	WORKLOAD_XML_IPERF:  WORKLOAD_JSON_IPERF,
	WORKLOAD_XML_VOIPRX: WORKLOAD_JSON_VOIPRX,
}

// find looks for an item in a slice.
// Reference: https://golangcode.com/check-if-element-exists-in-slice/
func find(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

// TestParseFlowOpOptions tests that ParseFlowOpOptions can successfully
// parse options given as space-separated key=value pairs, raising an error
// if a bad input is given.
func TestParseFlowOpOptions(t *testing.T) {
	var expected map[string]string = map[string]string{
		"key":                 "value",
		"hey_there":           "hi_mom",
		"this_is_another_KEY": "THIS_IS_aNoThEr_Value",
	}
	var inputStr string = "key=value hey_there=hi_mom this_is_another_KEY=THIS_IS_aNoThEr_Value"

	options, err := ParseFlowOpOptions(inputStr)

	if err != nil {
		t.Errorf("Got error while parsing known-good option string: %s", err)
	}

	var parsedKeys []string = []string{}
	for key, val := range options {
		expected_val, ok := expected[key]

		if !ok {
			t.Errorf(
				"Got unknown key in parsed output: %s", key,
			)
		}
		parsedKeys = append(parsedKeys, key)
		if expected_val != val {
			t.Errorf(
				"Expected value does not meet given value. Expected %s, got %s",
				expected_val, val,
			)
		}
	}

	for key := range expected {
		ok := find(parsedKeys, key)
		if !ok {
			t.Errorf(
				"Expected the following key in parsed output, but it wasn't found: %s", key,
			)
		}
	}
}

// TestParseWorkloadXML tries to parse sample profiles given by Uperf
// into JSON, testing if it matches prepared JSON versions of the profiles.
func TestParseWorkloadXML(t *testing.T) {
	for workload_xml, workload_json := range WORKLOAD_XML_TO_JSON_TESTS {
		var unmarshalled_json Profile = Profile{}

		err := json.Unmarshal([]byte(workload_json), &unmarshalled_json)
		if err != nil {
			t.Errorf("Error while trying to unmarshal workload JSON: %s", err)
		}

		_unmarshalled_xml, err := ParseWorkloadXML([]byte(workload_xml))
		if err != nil {
			t.Errorf("Error while trying to unmarshal workload XML: %s", err)
		}
		unmarshalled_xml := *_unmarshalled_xml

		if !reflect.DeepEqual(unmarshalled_json, unmarshalled_xml) {
			t.Errorf(
				"Expected json is not equal to parsed XML:\n json:\n%s\n xml:\n%s\n",
				unmarshalled_json, unmarshalled_xml,
			)
		}
	}
}
