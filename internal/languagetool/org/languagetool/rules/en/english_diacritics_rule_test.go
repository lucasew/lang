package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestEnglishDiacriticsRule(t *testing.T) {
	rule := NewEnglishDiacriticsRule(nil)
	// Java example: blase → blasé
	matches := rule.Match(languagetool.AnalyzePlain("He was quite blase about it."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "blasé", matches[0].GetSuggestedReplacements()[0])
}

func TestSimpleReplaceProfanityRule(t *testing.T) {
	rule := NewSimpleReplaceProfanityRule(nil)
	matches := rule.Match(languagetool.AnalyzePlain("That is an arsehole."))
	require.Equal(t, 1, len(matches))
	require.Contains(t, matches[0].GetMessage(), "offensive")
}
