package exporters

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/learnitall/gobench/define"
)

var HEALTHCHECK_RESPONSE_STR string = `{
	"name" : "587ecbdd1d24",
	"cluster_name" : "docker-cluster",
	"cluster_uuid" : "w_bA02yGQXq5oiWUSchEqg",
	"version" : {
	  "number" : "7.17.0",
	  "build_flavor" : "default",
	  "build_type" : "docker",
	  "build_hash" : "bee86328705acaa9a6daede7140defd4d9ec56bd",
	  "build_date" : "2022-01-28T08:36:04.875279988Z",
	  "build_snapshot" : false,
	  "lucene_version" : "8.11.1",
	  "minimum_wire_compatibility_version" : "6.8.0",
	  "minimum_index_compatibility_version" : "6.0.0-beta1"
	},
	"tagline" : "You Know, for Search"
  }`

func TestElasticProductHeaderIsInjected(t *testing.T) {
	testServer := httptest.NewTLSServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, HEALTHCHECK_RESPONSE_STR)
			},
		),
	)
	defer testServer.Close()

	testServerURL, err := url.Parse(testServer.URL)
	if err != nil {
		t.Fatalf("Unable to get url of test server: %s", err)
	}

	ctx := define.GetConfig()
	ctx.ElasticsearchURL = testServerURL.String()
	ctx.ElasticsearchSkipVerify = true
	ctx.ElasticsearchInjectProductHeader = false
	es := ElasticsearchExporter{}
	es.Setup(ctx)
	err = es.Healthcheck()

	if err == nil {
		t.Error(
			"Expected error during healthcheck (no X-Elastic-Product header injected), instead got success",
		)
	}

	ctx.ElasticsearchInjectProductHeader = true
	es = ElasticsearchExporter{}
	es.Setup(ctx)
	err = es.Healthcheck()

	if err != nil {
		t.Errorf(
			"Failed healthcheck when trying to inject X-Elastic-Product header: %s", err,
		)
	}
}

func TestElasticsearchExporterImplementsExporterInterface(t *testing.T) {
	var es interface{} = &ElasticsearchExporter{}
	_, ok := es.(define.Exporterable)

	// var _ define.Exporterable = &ElasticsearchExporter{}

	if !ok {
		t.Errorf(
			"ElasticsearchExporter failed Exporterable type assertion",
		)
	}
}
