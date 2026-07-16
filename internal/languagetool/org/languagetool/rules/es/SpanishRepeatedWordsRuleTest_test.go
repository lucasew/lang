package es

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSpanishRepeatedWordsRule(t *testing.T) {
	rule := NewSpanishRepeatedWordsRule(nil)
	sents := []*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("Voy a sugerir algo."),
		languagetool.AnalyzePlain("Puedo sugerir otra cosa."),
	}
	require.Equal(t, 1, len(rule.MatchList(sents)))
}
