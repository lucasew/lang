package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestStyleTooOftenUsedNounRule(t *testing.T) {
	rule := NewStyleTooOftenUsedNounRule(nil)
	// "Haus" appears twice mid-sentence as capitalized nouns
	sents := languagetool.SplitAndAnalyze("Er sah das Haus am See. Dann kaufte er das Haus in der Stadt.")
	matches := rule.MatchList(sents)
	require.GreaterOrEqual(t, len(matches), 2)
	// no mid-sentence noun repeat
	sents2 := languagetool.SplitAndAnalyze("Er sah das Haus am See. Dann kaufte er die Villa in der Stadt.")
	require.Equal(t, 0, len(rule.MatchList(sents2)))
}
