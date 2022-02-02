// context.go provides items which help define the current runtime context,
// such as configurations and shared objects.
package define

import (
	"sync"
)

// Contextable defines methods needed for the Context object to
// implement the singleton pattern.
// Reference: https://refactoring.guru/design-patterns/singleton/go/example
type Contextable interface {
	Get() Contextable
}

// Context defines common runtime objects which need to be accessed by
// other objects within gobench.
// It uses a singleton pattern.
type Context struct {
	Verbose            bool
	RunID              string
	ElasticsearchURL   string
	ElasticsearchIndex string
}

var contextLock = &sync.Mutex{}
var contextInstance *Context

// Get returns the current Context instance.
func GetContext() *Context {
	if contextInstance == nil {
		contextLock.Lock()
		defer contextLock.Unlock()
		if contextInstance == nil {
			contextInstance = &Context{}
		}
	}
	return contextInstance
}
