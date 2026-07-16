package ca

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestReplaceOperationNamesRule(t *testing.T) {
	rule := NewReplaceOperationNamesRule(nil)
	matches := rule.Match(languagetool.AnalyzePlain("procés d'encriptat de dades"))
	require.Equal(t, 1, len(matches))
	require.Contains(t, matches[0].GetSuggestedReplacements(), "encriptació")
}

func TestSimpleReplaceDNVColloquialRule(t *testing.T) {
	rule := NewSimpleReplaceDNVColloquialRule(nil)
	matches := rule.Match(languagetool.AnalyzePlain("vol acaminar al parc"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "caminar", matches[0].GetSuggestedReplacements()[0])
}
