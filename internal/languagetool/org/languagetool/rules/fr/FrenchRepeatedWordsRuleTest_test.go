package fr

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestFrenchRepeatedWordsRule(t *testing.T) {
	rule := NewFrenchRepeatedWordsRule(nil)
	// maintenant is in synonyms
	sents := []*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("Je le fais maintenant."),
		languagetool.AnalyzePlain("Et maintenant j'attends."),
	}
	matches := rule.MatchList(sents)
	require.Equal(t, 1, len(matches))
}
