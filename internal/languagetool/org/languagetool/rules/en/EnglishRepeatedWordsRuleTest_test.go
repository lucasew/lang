package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestEnglishRepeatedWordsRule(t *testing.T) {
	rule := NewEnglishRepeatedWordsRule(nil)
	// need appears twice across two period-ended sentences
	sents := []*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("I need help."),
		languagetool.AnalyzePlain("I still need time."),
	}
	matches := rule.MatchList(sents)
	require.Equal(t, 1, len(matches))
	require.Contains(t, matches[0].GetSuggestedReplacements(), "require")
}
