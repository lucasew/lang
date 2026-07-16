package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestStyleRepeatedVeryShortSentences(t *testing.T) {
	rule := NewStyleRepeatedVeryShortSentences(nil)
	// three short sentences in a row (Java example)
	sents := []*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("Das Auto kam näher."),
		languagetool.AnalyzePlain("Der Hund schlief."),
		languagetool.AnalyzePlain("Die Reifen quietschten."),
	}
	// "Der Hund schlief." is very short; others may be borderline depending on tokenization
	// Force with minWords high enough for first sentence too
	rule.MinWords = 5
	matches := rule.MatchList(sents)
	require.GreaterOrEqual(t, len(matches), 3)

	// long sentence breaks the streak
	sents2 := []*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("Er lief."),
		languagetool.AnalyzePlain("Sie rief."),
		languagetool.AnalyzePlain("Dieser deutlich längere Satz unterbricht die Serie von kurzen Sätzen hier."),
		languagetool.AnalyzePlain("Er ging."),
	}
	require.Equal(t, 0, len(rule.MatchList(sents2)))
}
