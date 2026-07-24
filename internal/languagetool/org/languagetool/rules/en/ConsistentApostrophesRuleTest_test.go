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

// Java ConsistentApostrophesRule: setDefaultTempOff, apostrophe URL, example pair, minParagraph -1.
func TestConsistentApostrophesRule_Metadata(t *testing.T) {
	rule := NewConsistentApostrophesRule(nil)
	require.Equal(t, "EN_CONSISTENT_APOS", rule.GetID())
	require.Contains(t, rule.GetDescription(), "apostrophes")
	require.Contains(t, rule.GetURL(), "what-is-an-apostrophe")
	require.True(t, rule.IsDefaultTempOff())
	require.Equal(t, -1, rule.MinToCheckParagraph())
	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "It's nice, but it <marker>doesn’t</marker> work.", inc[0].GetExample())
	require.Equal(t, []string{"doesn't"}, inc[0].GetCorrections())
	require.Equal(t, "It's nice, but it <marker>doesn't</marker> work.", rule.GetCorrectExamples()[0].GetExample())
}

func formatOne(s []string) string {
	if len(s) == 0 {
		return "[]"
	}
	return "[" + s[0] + "]"
}
