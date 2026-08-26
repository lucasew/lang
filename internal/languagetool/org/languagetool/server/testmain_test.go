package server

import (
	"os"
	"testing"
)

// Server unit tests build a Pipeline per request; loading multi-MB official
// grammar each time is too slow for the API surface suite. Production default
// remains UseUpstreamGrammar on (Java getRuleFileNames). Opt out here; tests
// that need official grammar set LANG_USE_UPSTREAM_GRAMMAR=1 explicitly.
func TestMain(m *testing.M) {
	if os.Getenv("LANG_USE_UPSTREAM_GRAMMAR") == "" {
		_ = os.Setenv("LANG_USE_UPSTREAM_GRAMMAR", "0")
	}
	os.Exit(m.Run())
}
