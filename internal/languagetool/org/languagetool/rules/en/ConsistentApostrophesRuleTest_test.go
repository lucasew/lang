package en

// Twin of languagetool-language-modules/en/src/test/java/org/languagetool/rules/en/ConsistentApostrophesRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestConsistentApostrophesRule_Rule(t *testing.T) {
	rule := NewConsistentApostrophesRule(nil)

	// One continuous analysis (EnglishWordTokenizer) matches Java lt.analyzeText positions.
	one := func(s string) []*languagetool.AnalyzedSentence {
		return []*languagetool.AnalyzedSentence{AnalyzeEnglishPlain(s)}
	}
	matches := rule.MatchList(one("It's a nice idea. But it doesn’t work."))
	require.Equal(t, 2, len(matches))
	require.Equal(t, 2, matches[0].GetFromPos())
	require.Equal(t, 4, matches[0].GetToPos())
	require.Equal(t, "[’s]", formatOne(matches[0].GetSuggestedReplacements()))
	require.Equal(t, 29, matches[1].GetFromPos())
	require.Equal(t, 32, matches[1].GetToPos())
	require.Equal(t, "[n't]", formatOne(matches[1].GetSuggestedReplacements()))

	require.Equal(t, 2, len(rule.MatchList(one("It’s a nice idea. But it doesn't work."))))
	require.Equal(t, 0, len(rule.MatchList(one("It's a nice idea. But it doesn't work."))))
	require.Equal(t, 0, len(rule.MatchList(one("It’s a nice idea. But it doesn’t work."))))
}

func formatOne(s []string) string {
	if len(s) == 0 {
		return "[]"
	}
	return "[" + s[0] + "]"
}
