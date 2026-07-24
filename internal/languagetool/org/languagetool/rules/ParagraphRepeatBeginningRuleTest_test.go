package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestParagraphRepeatBeginningRule(t *testing.T) {
	rule := NewParagraphRepeatBeginningRule(nil)
	// Next sentence text starts with newline → paragraph boundary
	sents := []*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("Wiederholung am Anfang."),
		languagetool.AnalyzePlain("\nWiederholung am Ende."),
	}
	matches := rule.MatchList(sents)
	require.Equal(t, 2, len(matches), "expected matches for repeated paragraph start, got %d", len(matches))
}
