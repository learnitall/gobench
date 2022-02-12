package exporters

import (
	"github.com/learnitall/gobench/define"
	"testing"
)

// TestDummyExporterImplementsExporterInterface does a quick check to make sure
// that the ChainExporter can successfully be type asserted as a define.Exporterable.
func TestDummyExporterImplementsExporterInterface(t *testing.T) {
	var ed interface{} = &DummyExporter{}
	_, ok := ed.(define.Exporterable)

	// Can use this line to help debug problems within IDE
	// var _ define.Exporterable = &DummyExporter{}

	if !ok {
		t.Errorf(
			"DummyExporter failed Exporterable type assertion",
		)
	}
}
