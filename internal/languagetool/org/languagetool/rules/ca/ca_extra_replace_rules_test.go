package ca

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestReplaceOperationNamesRule(t *testing.T) {
	rule := NewReplaceOperationNamesRule(nil)
	// "d'encriptat" needs Catalan tokenizer + prev SPS00 (via opNamesTagWord helper path)
	matches := rule.Match(analyzeOpNames("procés d'encriptat de dades"))
	require.Equal(t, 1, len(matches))
	require.Contains(t, matches[0].GetSuggestedReplacements(), "encriptació")
}

func TestSimpleReplaceDNVColloquialRule(t *testing.T) {
	rule := NewSimpleReplaceDNVColloquialRule(nil)
	// lemma path: inject acaminar lemma (Java Catalan tagger)
	matches := rule.Match(analyzeCALemma("vol acaminar al parc", map[string]languagetool.TokenTag{
		"acaminar": {POS: "VMN0000", Lemma: "acaminar"},
	}))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "caminar", matches[0].GetSuggestedReplacements()[0])
	// without lemma: fail closed
	require.Empty(t, rule.Match(languagetool.AnalyzePlain("vol acaminar al parc")))
}
