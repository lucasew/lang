package ca

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestCatalanRepeatedWordsRule(t *testing.T) {
	rule := NewCatalanRepeatedWordsRule(nil)
	sents := []*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("Vull suggerir una idea."),
		languagetool.AnalyzePlain("Puc suggerir una altra."),
	}
	require.Equal(t, 1, len(rule.MatchList(sents)))
}
