// elasticsearch.go implements the ElasticsearchExporter object, which is used to
// export benchmark results to ElasticSearch.
package exporters

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esutil"
	"github.com/learnitall/gobench/define"
	"github.com/rs/zerolog/log"
	"github.com/valyala/fasthttp"
)

// fasthttpTransport replaces elasticsearch's http client, based on net/http,
// with the http client provided by github.com/valyala/fasthttp
// Reference: https://github.com/elastic/go-elasticsearch/blob/main/_examples/fasthttp/fasthttp.go
type fasthttpTransport struct {
	_client              *fasthttp.Client
	_injectProductHeader bool
}

// copyRequest converts a http.Request to fasthttp.Request
func (t *fasthttpTransport) copyRequest(dst *fasthttp.Request, src *http.Request) *fasthttp.Request {
	if src.Method == "GET" && src.Body != nil {
		src.Method = "POST"
	}

	dst.SetHost(src.Host)
	dst.SetRequestURI(src.URL.String())

	dst.Header.SetRequestURI(src.URL.String())
	dst.Header.SetMethod(src.Method)

	for k, vv := range src.Header {
		for _, v := range vv {
			dst.Header.Set(k, v)
		}
	}

	if src.Body != nil {
		dst.SetBodyStream(src.Body, -1)
	}

	return dst
}

// copyResponse converts a fasthttp.Response to a http.Response
func (t *fasthttpTransport) copyResponse(dst *http.Response, src *fasthttp.Response) *http.Response {
	dst.StatusCode = src.StatusCode()

	src.Header.VisitAll(func(k, v []byte) {
		dst.Header.Set(string(k), string(v))
	})

	// https://towardsaws.com/elasticsearch-the-server-is-not-a-supported-distribution-of-elasticsearch-252abc1bd92
	if t._injectProductHeader {
		dst.Header.Set("X-Elastic-Product", "Elasticsearch")
	}

	// Cast to a string to make a copy seeing as src.Body() won't
	// be valid after the response is released back to the pool (fasthttp.ReleaseResponse).
	dst.Body = ioutil.NopCloser(strings.NewReader(string(src.Body())))

	return dst
}

// RoundTrip performs the request and returns a response or error
func (t *fasthttpTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	freq := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(freq)

	fres := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(fres)

	t.copyRequest(freq, req)

	err := t._client.Do(freq, fres)
	if err != nil {
		return nil, err
	}

	res := &http.Response{Header: make(http.Header)}
	t.copyResponse(res, fres)

	return res, nil
}

// ElasticsearchExporter is used to export benchmark results into Elasticsearch
type ElasticsearchExporter struct {
	cfg         *elasticsearch.Config
	client      *elasticsearch.Client
	bulkIndexer *esutil.BulkIndexer
	bulkCfg     *esutil.BulkIndexerConfig
	index       string
	clusterInfo map[string]interface{}
}

func (es *ElasticsearchExporter) Setup(cfg *define.Config) error {
	fasthttpClient := fasthttp.Client{
		TLSConfig: &tls.Config{
			InsecureSkipVerify: cfg.ElasticsearchSkipVerify,
		},
	}
	esCfg := elasticsearch.Config{
		Addresses: []string{
			cfg.ElasticsearchURL,
		},
		Transport: &fasthttpTransport{
			_client:              &fasthttpClient,
			_injectProductHeader: cfg.ElasticsearchInjectProductHeader,
		},
		// These options are references from
		// https://github.com/elastic/go-elasticsearch/blob/main/_examples/bulk/indexer.go
		RetryOnStatus: []int{502, 503, 504, 429},
		MaxRetries:    5,
		RetryBackoff: func(i int) time.Duration {
			// use binary exponential backoff
			return time.Duration(
				math.Floor(
					math.Pow(2, float64(i)),
				),
			) * time.Second
		},
	}
	es.cfg = &esCfg

	client, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Unable to create new client for ElasticsearchExporter.")
		return err
	}
	es.client = client
	es.index = cfg.ElasticsearchIndex

	// https://github.com/elastic/go-elasticsearch/blob/main/_examples/bulk/indexer.go
	bulkCfg := esutil.BulkIndexerConfig{
		Index:  es.index,
		Client: es.client,
		// Just using sane defaults for now, can be modified as needed
		NumWorkers:    3,
		FlushInterval: 10 * time.Second,
		FlushBytes:    1e+6, // ~ 1024 KiB or 1 MiB
		OnError: func(ctx context.Context, err error) {
			log.Warn().
				Err(err).
				Msg("Received error while indexing item through the bulk indexer")
		},
	}
	es.bulkCfg = &bulkCfg

	bi, err := esutil.NewBulkIndexer(bulkCfg)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Unable to create new bulk indexer for ElasticsearchExporter.")
	}
	es.bulkIndexer = &bi

	log.Info().
		Interface("cfg", cfg).
		Msg("Created new ElasticsearchExporter")

	return nil
}

func (es *ElasticsearchExporter) Healthcheck() error {
	_healthcheck_failed_str := "Healcheck failed for ElasticsearchExporter"

	if es.client == nil {
		err := errors.New(
			"Healthcheck called on ElasticsearchExporter which hasn't been setup yet",
		)
		log.Warn().
			Err(err).
			Msg("Was Setup called on the ElasticsearchExporter?")
		return err
	}

	res, err := es.client.Info()
	if err != nil {
		log.Warn().
			Err(err).
			Msg(
				fmt.Sprintf(
					"%s, go-level error when getting cluster info",
					_healthcheck_failed_str,
				),
			)
		return err
	}
	defer res.Body.Close()

	if res.IsError() {
		log.Warn().
			Str("response", res.String()).
			Msg("Healthcheck failed for ElasticsearchExporter, es-level error when getting cluster info")
		return fmt.Errorf(
			"%s: %s",
			_healthcheck_failed_str,
			res.String(),
		)
	}

	if err := json.NewDecoder(res.Body).Decode(&es.clusterInfo); err != nil {
		err_str := fmt.Sprintf(
			"%s, %s",
			_healthcheck_failed_str,
			"unable to decode response body",
		)
		log.Warn().
			Interface("response_body", res.Body).
			Msg(err_str)
		return fmt.Errorf(
			"%s: %s",
			err_str, res.Body,
		)
	}
	log.Info().
		Str(
			"client_version",
			elasticsearch.Version,
		).
		Str(
			"server_version",
			fmt.Sprintf(
				"%s",
				es.clusterInfo["version"].(map[string]interface{})["number"],
			),
		).
		Msg("Healcheck for ElasticsearchExporter successful!")
	return nil
}

func (es *ElasticsearchExporter) Teardown() error {
	indexer := *es.bulkIndexer
	if err := indexer.Close(context.Background()); err != nil {
		log.Error().
			Err(err).
			Msg("Unexpected error while closing out bulk indexer.")
		return err
	}
	return nil
}

func (es *ElasticsearchExporter) Marshal(payload interface{}) ([]byte, error) {
	return json.Marshal(payload)
}

func (es *ElasticsearchExporter) Export(payload []byte) error {
	indexer := *es.bulkIndexer
	err := indexer.Add(
		context.Background(),
		esutil.BulkIndexerItem{
			Action: "index",
			Body:   bytes.NewReader(payload),
			OnFailure: func(
				c context.Context,
				bii esutil.BulkIndexerItem,
				biri esutil.BulkIndexerResponseItem,
				e error,
			) {
				if e != nil {
					log.Warn().
						Err(e).
						Msg("go-level error while indexing with bulk indexer.")
				} else {
					log.Warn().
						Str("error_type", biri.Error.Type).
						Str("error_reason", biri.Error.Reason).
						Msg("es-level error while indexing with bulk indexer.")
				}
			},
		},
	)

	if err != nil {
		log.Error().
			Err(err).
			RawJSON("payload", payload).
			Msg("Unexpected error while adding item to bulk indexer")
		return err
	}

	return nil
}
