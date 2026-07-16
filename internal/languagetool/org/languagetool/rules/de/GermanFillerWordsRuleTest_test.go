package de

// Twin of GermanFillerWordsRuleTest (minPercent=0 surface mode).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestGermanFillerWordsRule_Rule(t *testing.T) {
	rule := NewGermanFillerWordsRule(nil)
	require.Equal(t, "FILLER_WORDS_DE", rule.GetID())
	// With minPercent=0 every filler is reported
	matches := rule.Match(languagetool.AnalyzePlain("Der Satz enthält augenscheinlich ein Füllwort."))
	require.Equal(t, 1, len(matches))
	matches = rule.Match(languagetool.AnalyzePlain("Der Satz enthält augenscheinlich relativ viele Füllwörter."))
	require.Equal(t, 2, len(matches))
}
