package en

// Twin of languagetool-language-modules/en/src/test/java/org/languagetool/rules/en/EnglishUnpairedBracketsRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestEnglishUnpairedBracketsRule_Rule(t *testing.T) {
	rule := NewEnglishUnpairedBracketsRule(nil)
	matchN := func(s string) int {
		return len(rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain(s)}))
	}
	require.Equal(t, 0, matchN("(This is a test sentence)."))
	require.Equal(t, 0, matchN("This is no smiley: (some more text)"))
	require.Equal(t, 0, matchN("This is a sentence with a smiley :)"))
	require.Equal(t, 0, matchN("This is a sentence with a smiley :("))
	require.Equal(t, 0, matchN("This is a sentence with a smiley :-)"))
	require.Equal(t, 0, matchN("This is a [test] sentence..."))
	require.Equal(t, 0, matchN("(([20] [20] [20]))"))
	// incorrect
	require.Equal(t, 1, matchN("This is a test sentence)."))
	require.Equal(t, 1, matchN("(This is a test sentence."))
}

func TestEnglishUnpairedBracketsRule_MultipleSentences(t *testing.T) {
	// Surface twin: single-sentence stack for now; multi-sentence text paths differ.
	rule := NewEnglishUnpairedBracketsRule(nil)
	sents := []*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("This is correct (yes)."),
		languagetool.AnalyzePlain("Still fine."),
	}
	require.Equal(t, 0, len(rule.MatchList(sents)))
}

// Java EnglishUnpairedBracketsRule: EN_UNPAIRED_BRACKETS, parentheses URL, example pair.
func TestEnglishUnpairedBracketsRule_Metadata(t *testing.T) {
	rule := NewEnglishUnpairedBracketsRule(nil)
	require.Equal(t, "EN_UNPAIRED_BRACKETS", rule.GetID())
	require.Contains(t, rule.GetURL(), "what-are-parentheses")
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "PUNCTUATION", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSTypographical, rule.GetLocQualityIssueType())
	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "He lived in a <marker>(</marker>large house.", inc[0].GetExample())
	// Java Rule.addExamplePair uses first fixed <marker> span as correction
	require.Equal(t, []string{"("}, inc[0].GetCorrections())
	require.Contains(t, rule.GetCorrectExamples()[0].GetExample(), "large")
}
