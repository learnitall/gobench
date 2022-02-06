// export.go defines items relevant to the export of benchmark data.
package define

// Exporterable defines methods needed by concrete Exporter objects.
// It is assumed that Exporterable objects are created with the intention
// of being added into the current runtime's Config.
type Exporterable interface {
	// Setup takes current Config and gets the Exporter ready to export data.
	// As an example, this could create a goroutine and some channels to
	// perform async export in the background.
	Setup(*Config) error
	// Teardown takes in current Config and closes the Exporter down.
	Teardown(*Config) error
	// Healthcheck ensures that the exporter is ready to function.
	Healthcheck() error
	// Marshal prepares the given object to be exported by marshaling it into
	// a byte string.
	Marshal(interface{}) ([]byte, error)
	// Export takes the given byte string (assumed to be marshaled) and exports
	// it. Differenc exporters can choose whether to make this async or sync.
	Export([]byte) error
}
