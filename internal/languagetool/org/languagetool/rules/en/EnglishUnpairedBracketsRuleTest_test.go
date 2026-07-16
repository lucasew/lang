package en

// Twin of languagetool-language-modules/en/src/test/java/org/languagetool/rules/en/EnglishUnpairedBracketsRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
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
