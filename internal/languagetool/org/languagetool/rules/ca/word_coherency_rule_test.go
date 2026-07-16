package ca

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestWordCoherencyRule(t *testing.T) {
	rule := NewWordCoherencyRule(nil)
	// Java example: pesebre … pessebre
	matches := rule.MatchList([]*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("Un pesebre ací i un altre pessebre allà."),
	})
	require.NotEmpty(t, matches)
}

func TestWordCoherencyValencianRule(t *testing.T) {
	rule := NewWordCoherencyValencianRule(nil)
	// Java example: Este … aquest
	matches := rule.MatchList([]*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("Este home d'ací parla amb aquest altre ací."),
	})
	require.NotEmpty(t, matches)
}
