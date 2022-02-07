// config.go provides items which help define the current runtime configuration.
package define

import (
	"sync"
)

// Config defines common runtime objects which need to be accessed by
// other objects within gobench.
// It uses a singleton pattern.
// Reference: https://refactoring.guru/design-patterns/singleton/go/example
type Config struct {
	Verbose                          bool
	RunID                            string
	ElasticsearchURL                 string
	ElasticsearchIndex               string
	ElasticsearchSkipVerify          bool
	ElasticsearchInjectProductHeader bool
}

var configLock = &sync.Mutex{}
var configInstance *Config

// Get returns the current Config instance.
func GetConfig() *Config {
	if configInstance == nil {
		configLock.Lock()
		defer configLock.Unlock()
		if configInstance == nil {
			configInstance = &Config{}
		}
	}
	return configInstance
}
