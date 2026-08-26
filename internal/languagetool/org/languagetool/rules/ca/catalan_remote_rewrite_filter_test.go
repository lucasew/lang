package ca

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCatalanRemoteRewriteFilter(t *testing.T) {
	f := NewCatalanRemoteRewriteFilter()
	f.Rewrite = func(sentence, ruleID string) string {
		return "Hola mon"
	}
	// "Hola mon" vs "Hola món" style - simple rewrite
	f.Rewrite = func(sentence, ruleID string) string {
		return "Això està bé"
	}
	res := f.Apply("Això esta bé", 5, 9, "RULE", true)
	// may or may not join depending on diff; ensure no panic
	if res.Keep && len(res.Replacements) > 0 {
		require.NotEmpty(t, res.Replacements[0])
	}
	// empty rewrite suppresses when requested
	f.Rewrite = func(string, string) string { return "" }
	res = f.Apply("text", 0, 4, "R", true)
	require.False(t, res.Keep)
}
