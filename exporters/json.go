package exporters

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/learnitall/gobench/define"
)

type JsonExporter struct {
	documents []string
}

// Setup creates a new array to hold json documents to-be-printed.
func (je *JsonExporter) Setup(cfg *define.Config) error {
	je.documents = []string{}
	return nil
}

// Healthcheck returns nil, no healthcheck needs to be performed for the json exporter.
func (je *JsonExporter) Healthcheck() error {
	return nil
}

// Marshal turns the given payload into pretty-formatted json.
func (je *JsonExporter) Marshal(payload interface{}) ([]byte, error) {
	return json.MarshalIndent(payload, "", "    ")
}

// Export saves the given document, which will be printed when Teardown is called.
func (je *JsonExporter) Export(payload []byte) error {
	je.documents = append(je.documents, string(payload))
	return nil
}

// Teardown joins all the given documents to export into an array and prints the result.
func (je *JsonExporter) Teardown() error {
	_, err := fmt.Printf("[%s]", strings.Join(je.documents, ","))
	return err
}
