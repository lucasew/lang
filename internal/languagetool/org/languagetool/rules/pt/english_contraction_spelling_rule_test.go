package pt

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestEnglishContractionSpellingRule(t *testing.T) {
	rule := NewEnglishContractionSpellingRule(nil)
	// Java example: whats → what's
	matches := rule.Match(languagetool.AnalyzePlain("Ele adorava assistir whats cooking."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "what's", matches[0].GetSuggestedReplacements()[0])
}
