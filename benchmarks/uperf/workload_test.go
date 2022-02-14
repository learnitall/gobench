//go:build uperf_test
// +build uperf_test

package uperf

import (
	"encoding/json"
	"os"
	"testing"
)

// These values need to be synced within the above strings.
var ENV_SUBST_VARS map[string]string = map[string]string{
	"h":     "127.0.0.1",
	"h2":    "myhost.custom",
	"proto": "tcp",
	"nthr":  "10",
}

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
        <flowop type="connect" options="remotehost=$h2 
		protocol=udp"/>
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
			"nthreads": 200,
			"XMLName": "group",
			"transactions": [
				{
					"iterations": 1,
					"flowops": [
						{
						    "type": "connect",
							"options": "remotehost=127.0.0.1 protocol=udp"
						}
					]
				},
				{
					"durationSeconds": 120,
					"flowops": [
						{
							"type": "read",
							"options": "size=64"
						}
					]
				},
				{
					"iterations": 1,
					"flowops": [
						{
							"type": "disconnect"
						}
					]
				}
			]
		},
		{
			"nthreads": 200,
			"transactions": [
				{
					"iterations": 1,
					"flowops": [
						{
							"type": "connect",
							"options": "remotehost=myhost.custom protocol=udp"
						}
					]
				},
				{
					"durationSeconds": 120,
					"flowops": [
						{
							"type": "read",
							"options": "size=64"
						}
					]
				},
				{
					"iterations": 1,
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
        <flowop type="connect" options="remotehost=myhost.custom 
		protocol=udp"/>
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
			"nthreads": 10,
			"transactions": [
				{
					"iterations": 1,
					"flowops": [
						{
						    "type": "connect",
						    "options": "remotehost=127.0.0.1 protocol=tcp wndsz=50k tcp_nodelay"
						}
					]
				},
				{
					"durationSeconds": 30,
					"flowops": [
						{
							"type": "write",
							"options": "count=10 size=8k"
						}
					]
				},
				{
					"iterations": 1,
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

// TestParseWorkloadXML tries to parse sample profiles given by Uperf
// into JSON, testing if it matches prepared JSON versions of the profiles.
func TestParseWorkloadXML(t *testing.T) {
	for env_key, env_value := range ENV_SUBST_VARS {
		os.Setenv(env_key, env_value)
	}
	for workload_xml, workload_json := range WORKLOAD_XML_TO_JSON_TESTS {
		var unmarshalled_json Profile = Profile{}

		err := json.Unmarshal([]byte(workload_json), &unmarshalled_json)
		if err != nil {
			t.Errorf("Error while trying to unmarshal workload JSON: %s", err)
		}

		env_workload_xml, err := PerformEnvSubst([]byte(workload_xml))
		if err != nil {
			t.Errorf("Error while trying to perform env substitutions: %s", err)
		}

		_unmarshalled_xml, err := ParseWorkloadXML(env_workload_xml)
		if err != nil {
			t.Errorf("Error while trying to unmarshal workload XML: %s", err)
		}
		unmarshalled_xml := *_unmarshalled_xml

		// To make the equal comparison, remarshal into equal json strings
		// This sets the indentation formatting to be on equal ground and takes
		// care of pointer defreference.
		remarshalled_xml, err := json.Marshal(unmarshalled_xml)
		if err != nil {
			t.Errorf("Error while trying to re-marshal workload xml: %s", err)
		}
		remarshalled_json, err := json.Marshal(unmarshalled_json)
		if err != nil {
			t.Errorf("Error while trying to re-marshal workload json: %s", err)
		}

		if string(remarshalled_xml) != string(remarshalled_json) {
			t.Errorf(
				"expected json is not equal to parsed XML:\n json:\n%+v\n xml:\n%+v\n",
				string(remarshalled_json), string(remarshalled_xml),
			)
		}
	}

	for env_key := range ENV_SUBST_VARS {
		os.Unsetenv(env_key)
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
