package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestGermanDoublePunctuationRule(t *testing.T) {
	rule := NewGermanDoublePunctuationRule(nil)
	require.Equal(t, "DE_DOUBLE_PUNCTUATION", rule.GetID())
	// Java example style: a. D..
	matches := rule.Match(languagetool.AnalyzePlain("Sein Vater ist Regierungsrat a. D.."))
	require.Equal(t, 1, len(matches))
	require.Contains(t, matches[0].GetMessage(), "Punkte")
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Sein Vater ist Regierungsrat a. D."))))
}
