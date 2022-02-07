package exporters

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/learnitall/gobench/define"
)

type MockPayload struct {
	Key string
}

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

var BULK_INDEX_RESPONSE_SUCCESS_STR string = `{
	"took": 5,
	"errors": false,
	"items": [
	   {
		  "index": {
			 "_index": "myIndex",
			 "_type": "_doc",
			 "_id": "1",
			 "_version": 1,
			 "result": "created",
			 "_shards": {
				"total": 2,
				"successful": 1,
				"failed": 0
			 },
			 "status": 201,
			 "_seq_no" : 0,
			 "_primary_term": 1
		  }
	   }
	]
}`

// getTestServer creates a new httptest.NewTLSServer for mocking ElasticSearch.
func getTestServer(
	t *testing.T, handler http.Handler,
) (*httptest.Server, *url.URL) {
	testServer := httptest.NewTLSServer(handler)

	testServerURL, err := url.Parse(testServer.URL)
	if err != nil {
		t.Fatalf("Unable to get url of test server: %s", err)
	}

	return testServer, testServerURL
}

// TestElasticProductHeaderIsInjected ensures that the injection of the X-Elastic-Product
// header by the http client given to the ElasticSearch library is functioning
// properly.
// When successful, the library will successfully perform a 'product check'.
// Reference: https://github.com/elastic/go-elasticsearch/blob/8134a159aafedf58af2780ebb3a30ec1938956f3/elasticsearch.go#L306
// This test also inadvertently tests if the ElasticsearchSkipVerify
// parameter works properly, as the local test server is setup with
// a self-signed certificate.
func TestElasticProductHeaderIsInjected(t *testing.T) {
	testServer, testServerURL := getTestServer(
		t,
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, HEALTHCHECK_RESPONSE_STR)
			},
		),
	)
	defer testServer.Close()

	cfg := define.GetConfig()
	cfg.ElasticsearchURL = testServerURL.String()
	cfg.ElasticsearchSkipVerify = true
	cfg.ElasticsearchInjectProductHeader = false
	es := ElasticsearchExporter{}
	es.Setup(cfg)
	err := es.Healthcheck()

	if err == nil {
		t.Error(
			"Expected error during healthcheck (no X-Elastic-Product header injected), instead got success",
		)
	}

	cfg.ElasticsearchInjectProductHeader = true
	es = ElasticsearchExporter{}
	es.Setup(cfg)
	err = es.Healthcheck()

	if err != nil {
		t.Errorf(
			"Failed healthcheck when trying to inject X-Elastic-Product header: %s", err,
		)
	}
}

// TestElasticsearchExporterImplementsExporterInterface does a quick check to make sure
// that the ElasticsearchExporter can successfully be type asserted as a define.Exporterable.
func TestElasticsearchExporterImplementsExporterInterface(t *testing.T) {
	var es interface{} = &ElasticsearchExporter{}
	_, ok := es.(define.Exporterable)

	// Can use this line to help debug problems within IDE
	// var _ define.Exporterable = &ElasticsearchExporter{}

	if !ok {
		t.Errorf(
			"ElasticsearchExporter failed Exporterable type assertion",
		)
	}
}

// TestElasticsearchExporterCanExportPayloads mocks an ElasticSearch server
// to determine if the Export method can successfully interact with it without
// any unexpected errors.
// This involves testing Setup, Marshal, Export and Teardown.
func TestElasticsearchExporterCanExportPayloads(t *testing.T) {
	testServer, testServerURL := getTestServer(
		t,
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, BULK_INDEX_RESPONSE_SUCCESS_STR)
				w.WriteHeader(http.StatusOK)
			},
		),
	)
	defer testServer.Close()

	cfg := define.GetConfig()
	cfg.ElasticsearchURL = testServerURL.String()
	// ensure this is synced with BULK_INDEX_RESPONSE_SUCCESS_STR
	cfg.ElasticsearchIndex = "myIndex"
	cfg.ElasticsearchSkipVerify = true
	cfg.ElasticsearchInjectProductHeader = true
	es := ElasticsearchExporter{}

	err := es.Setup(cfg)
	if err != nil {
		t.Errorf(
			"Got error during setup: %s", err,
		)
	}

	payload, err := es.Marshal(MockPayload{Key: "value"})
	if err != nil {
		t.Errorf(
			"Got error during marshal: %s", err,
		)
	}

	err = es.Export(payload)
	if err != nil {
		t.Errorf(
			"Received error while calling Export: %s", err,
		)
	}

	err = es.Teardown()
	if err != nil {
		t.Errorf(
			"Received error while calling Teardown: %s", err,
		)
	}

	stats := (*es.bulkIndexer).Stats()
	if stats.NumIndexed != 1 || stats.NumAdded != 1 {
		t.Errorf(
			"Expected bulk indexer stats to show 1 document added and indexed, instead got: %+v", stats,
		)
	}
}

// TestElasticsearchExporterFailsGracefullyOnExportFail makes sure that if the
// ElasticSearch server sends a bad response, the client will not error-out.
// This test goes about this by sending an InternalServerError back to the client, however
// further testing can be done by sending a mocked response which indicates
// a failure to the client.
// This test does not test the exponential back-off functionality of the
// bulkIndexer.
func TestElasticsearchExporterFailsGracefullyOnExportFail(t *testing.T) {
	testServer, testServerURL := getTestServer(
		t,
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
		),
	)
	defer testServer.Close()

	cfg := define.GetConfig()
	cfg.ElasticsearchURL = testServerURL.String()
	cfg.ElasticsearchIndex = "myIndex"
	cfg.ElasticsearchSkipVerify = true
	cfg.ElasticsearchInjectProductHeader = true
	es := ElasticsearchExporter{}

	err := es.Setup(cfg)
	if err != nil {
		t.Errorf(
			"Got error during setup: %s", err,
		)
	}

	payload, err := es.Marshal(MockPayload{Key: "value"})
	if err != nil {
		t.Errorf(
			"Got error during marshal: %s", err,
		)
	}

	err = es.Export(payload)
	if err != nil {
		t.Errorf(
			"Got error during export: %s", err,
		)
	}

	err = es.Teardown()
	if err != nil {
		t.Errorf(
			"Received error while calling Teardown: %s", err,
		)
	}

	stats := (*es.bulkIndexer).Stats()
	if stats.NumIndexed != 0 || stats.NumAdded != 1 {
		t.Errorf(
			"Expected bulk indexer stats to show 1 document added and failed to indexed, instead got: %+v", stats,
		)
	}
}
