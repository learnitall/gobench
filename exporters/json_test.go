package exporters

import (
	"testing"

	"github.com/learnitall/gobench/define"
)

// TestJsonExporterImplementsExporterInterface does a quick check to make sure
// that the JsonExporter can successfully be type asserted as a define.Exporterable.
func TestJsonExporterImplementsExporterInterface(t *testing.T) {
	var ej interface{} = &JsonExporter{}
	_, ok := ej.(define.Exporterable)

	// Can use this line to help debug problems within IDE
	// var _ define.Exporterable = &JsonExporter{}

	if !ok {
		t.Errorf(
			"JsonExporter failed Exporterable type assertion",
		)
	}
}
