package uperf

import (
	"encoding/json"
	"os"
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
var WORKLOAD_XML_VOIPRX_ENV_SUBST string = `<?xml version="1.0"?>
<profile name="VoIP Rx">
  <group nthreads="200">
      <transaction iterations="1">
        <flowop type="connect" options="remotehost=127.0.0.1 protocol=udp"/>
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
        <flowop type="connect" options="remotehost=myhost.custom protocol=udp"/>
      </transaction>
      <transaction duration="120s">
            <flowop type="read" options="size=64"/>
      </transaction>
      <transaction iterations="1">
            <flowop type="disconnect" />
      </transaction>
  </group>
</profile>`

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
var WORKLOAD_XML_IPERF_ENV_SUBST string = `<?xml version="1.0"?>
<profile name="iPERF">
  <group nthreads="10">
        <transaction iterations="1">
            <flowop type="connect" options="remotehost=127.0.0.1 protocol=tcp wndsz=50k  tcp_nodelay"/>
        </transaction>
        <transaction duration="30s">
            <flowop type="write" options="count=10 size=8k"/>
        </transaction>
        <transaction iterations="1">
            <flowop type="disconnect" />
        </transaction>
  </group>

</profile>`

var WORKLOAD_XML_TO_JSON_TESTS map[string]string = map[string]string{
	WORKLOAD_XML_IPERF:  WORKLOAD_JSON_IPERF,
	WORKLOAD_XML_VOIPRX: WORKLOAD_JSON_VOIPRX,
}

var WORKLOAD_XML_ENV_SUBST_TESTS map[string]string = map[string]string{
	WORKLOAD_XML_IPERF:  WORKLOAD_XML_IPERF_ENV_SUBST,
	WORKLOAD_XML_VOIPRX: WORKLOAD_XML_VOIPRX_ENV_SUBST,
}

// These values need to be synced within the above strings.
var ENV_SUBST_VARS map[string]string = map[string]string{
	"h":     "127.0.0.1",
	"h2":    "myhost.custom",
	"proto": "tcp",
	"nthr":  "10",
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

// TestPerformEnvSubst tries to replace environment variables within workload XML files,
// checking for errors if the environment variables aren't set and checking for
// success if they are.
// The following environment variables will be set for each workload xml:
// - h=127.0.0.1
// - h2=myhost.custom
// - proto=tcp
// - nthr=10
// See ENV_SUBST_VARS.
func TestPerformEnvSubst(t *testing.T) {
	for workload_xml, workload_xml_sub := range WORKLOAD_XML_ENV_SUBST_TESTS {
		workload_xml_bytes := []byte(workload_xml)
		workload_xml_sub_bytes := []byte(workload_xml_sub)

		result, err := PerformEnvSubst(workload_xml_bytes)
		if err == nil {
			t.Errorf(
				"Expected error when no env vars are set, instead got successful output: %s",
				result,
			)
		}

		for key, value := range ENV_SUBST_VARS {
			os.Setenv(key, value)
		}

		result, err = PerformEnvSubst(workload_xml_bytes)
		if err != nil {
			t.Errorf(
				"Got error while performing env subst on xml containing environment variables: %s",
				err,
			)
		}

		result_str := string(result)
		if result_str != workload_xml_sub {
			t.Errorf(
				"Got unexpected result while performing env subst on xml containing environment variables\nExpected:\n%s\nResult:\n%s",
				workload_xml_sub,
				result_str,
			)
		}

		for key := range ENV_SUBST_VARS {
			os.Unsetenv(key)
		}

		result, err = PerformEnvSubst(workload_xml_sub_bytes)
		if err != nil {
			t.Errorf(
				"Got error while performing env subst on xml not containing environment variables: %s",
				err,
			)
		}

		result_str = string(result)
		if result_str != workload_xml_sub {
			t.Errorf(
				"Got unexpected result while performing env subst on xml not containing environment variables\nExpected:\n%s\nResult:\n%s",
				workload_xml_sub,
				result_str,
			)
		}
	}
}
