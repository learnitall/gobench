package exporters

import (
	"github.com/learnitall/gobench/define"
)

// DummyExporter is an exporter that does nothing and returns no errors.
// Used as a place-holder when no exporters are configured.
type DummyExporter struct{}

// Setup creates a new array to hold json documents to-be-printed.
func (de *DummyExporter) Setup(cfg *define.Config) error {
	return nil
}

// Healthcheck returns nil, no healthcheck needs to be performed for the json exporter.
func (de *DummyExporter) Healthcheck() error {
	return nil
}

// Marshal turns the given payload into pretty-formatted json.
func (de *DummyExporter) Marshal(payload interface{}) ([]byte, error) {
	return []byte{}, nil
}

// Export saves the given document, which will be printed when Teardown is called.
func (de *DummyExporter) Export(payload []byte) error {
	return nil
}

// Teardown joins all the given documents to export into an array and prints the result.
func (de *DummyExporter) Teardown() error {
	return nil
}
