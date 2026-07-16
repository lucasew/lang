package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestNonSignificantVerbsRule(t *testing.T) {
	rule := NewNonSignificantVerbsRule(nil)
	// machte is non-significant
	matches := rule.Match(languagetool.AnalyzePlain("Er machte einen Kuchen."))
	require.Equal(t, 1, len(matches))
	// Angst exception with machen
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Das macht mir Angst."))))
}
