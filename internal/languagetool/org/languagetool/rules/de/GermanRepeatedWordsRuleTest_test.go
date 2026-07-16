package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestGermanRepeatedWordsRule(t *testing.T) {
	rule := NewGermanRepeatedWordsRule(nil)
	sents := []*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("Außerdem regnet es."),
		languagetool.AnalyzePlain("Außerdem ist es kalt."),
	}
	require.Equal(t, 1, len(rule.MatchList(sents)))
}
