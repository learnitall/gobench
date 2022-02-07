package define

import "testing"

// TestGetConfig tests that GetConfig always returns the same object,
// enforcing the single pattern. It does this by calling GetConfig,
// modifying the returned object, then calling GetConfig again to see if
// the modifications stuck.
func TestGetConfig(t *testing.T) {
	var ctx *Config = GetConfig()
	if ctx == nil {
		t.Fatal("First call to GetConfig() is nil, want Config Object.")
	}
	var verboseBefore bool = ctx.Verbose
	ctx.Verbose = !verboseBefore
	ctx = GetConfig()
	if ctx.Verbose == verboseBefore {
		t.Fatalf(
			"Expected changing Config.Verbose from %t to %t would stick across calls to GetConfig()",
			verboseBefore,
			!verboseBefore,
		)
	}
}
