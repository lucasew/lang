package de

import (
	"os"
	"testing"
)

// de unit tests register core packs frequently; full grammar.xml is multi-MB.
// Production default is UseUpstreamGrammar on; opt out for package tests.
// Explicit WireGermanUpstreamGrammar tests set the env themselves.
func TestMain(m *testing.M) {
	if os.Getenv("LANG_USE_UPSTREAM_GRAMMAR") == "" {
		_ = os.Setenv("LANG_USE_UPSTREAM_GRAMMAR", "0")
	}
	os.Exit(m.Run())
}
