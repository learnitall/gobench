package exporters

import (
	"testing"

	"github.com/learnitall/gobench/define"
)

// TestChainExporterImplementsExporterInterface does a quick check to make sure
// that the ChainExporter can successfully be type asserted as a define.Exporterable.
func TestChainExporterImplementsExporterInterface(t *testing.T) {
	var ec interface{} = &ChainExporter{}
	_, ok := ec.(define.Exporterable)

	// Can use this line to help debug problems within IDE
	// var _ define.Exporterable = &ChainExporter{}

	if !ok {
		t.Errorf(
			"ChainExporter failed Exporterable type assertion",
		)
	}
}
