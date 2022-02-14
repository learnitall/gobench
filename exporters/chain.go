package exporters

import (
	"github.com/learnitall/gobench/define"
)

// ChainExporter is used to allow for multiple exporters to function through one Expoerterable interface.
// It's important to initiate Exporters and marshalled appropriately
type ChainExporter struct {
	Exporters  []define.Exporterable
	Marshalled [][]byte
}

// doLoop loops through each Exporterable the ChainExporter is configured with and passes it to the given function.
// If an error is returned by the function, then the loop breaks and the error is returned.
func (ce *ChainExporter) doLoop(
	loopFunc func(define.Exporterable, int) error,
) error {
	for i, exporter := range ce.Exporters {
		err := loopFunc(exporter, i)
		if err != nil {
			return err
		}
	}
	return nil
}

// Setup calls the Setup method on each Exporterable the ChainExporter is configured with.
func (ce *ChainExporter) Setup(cfg *define.Config) error {
	return ce.doLoop(
		func(e define.Exporterable, i int) error {
			return e.Setup(cfg)
		},
	)
}

// Healthcheck calls the Healthcheck method on each Exporterable the ChainExporter is configured with.
func (ce *ChainExporter) Healthcheck() error {
	return ce.doLoop(
		func(e define.Exporterable, i int) error {
			return e.Healthcheck()
		},
	)
}

// Teardown calls the Teardown method on each Exporterable the ChainExporter is configured with.
func (ce *ChainExporter) Teardown() error {
	return ce.doLoop(
		func(e define.Exporterable, i int) error {
			return e.Teardown()
		},
	)
}

// Marshal calls the Marshal method on each Exporterable the ChainExporter is configured with.
// Rather than returning the results of each marshal, they are saved in a slice within the ChainExporter.
// An empty byte array is returned.
// The Exporter function will use these results while exporting payloads.
// This function is ripe for a runtime panic if the `marshalled` field is not
// preallocated to match the length of the number of configured exporters.
func (ce *ChainExporter) Marshal(payload interface{}) ([]byte, error) {
	err := ce.doLoop(
		func(e define.Exporterable, i int) error {
			marshalled, _err := e.Marshal(payload)
			if _err != nil {
				return _err
			}
			ce.Marshalled[i] = marshalled
			return nil
		},
	)
	return []byte{}, err
}

// Export calls the Export method on each Exporterable the ChainExporter is configured with, using the saved payloads from Marshal.
// This assumes that the ChainExporter's Marshal method has already been called.
// Otherwise, an out-of-bound slice error might be raised.
func (ce *ChainExporter) Export(payload []byte) error {
	return ce.doLoop(
		func(e define.Exporterable, i int) error {
			return e.Export(ce.Marshalled[i])
		},
	)
}
