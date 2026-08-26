package commandline

import (
	"os"
	"testing"
)

// Unit tests configure LT repeatedly; full grammar/style XML is multi-MB and
// multi-minute. Production UseUpstreamGrammar remains default-on (Java
// getRuleFileNames). Opt out for this package; tests that need official
// grammar set LANG_USE_UPSTREAM_GRAMMAR=1 explicitly (e.g. official_grammar_test).
func TestMain(m *testing.M) {
	if os.Getenv("LANG_USE_UPSTREAM_GRAMMAR") == "" {
		_ = os.Setenv("LANG_USE_UPSTREAM_GRAMMAR", "0")
	}
	os.Exit(m.Run())
}
