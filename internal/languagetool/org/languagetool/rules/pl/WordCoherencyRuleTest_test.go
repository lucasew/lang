package pl

// Twin of languagetool-language-modules/pl/src/test/java/org/languagetool/rules/pl/WordCoherencyRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestWordCoherencyRule_Rule(t *testing.T) {
	rule := NewWordCoherencyRule(nil)

	require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("To jest grejpfrut. Dobry grejpfrut."),
	})))
	require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("Lubię Twoje blefy. Blef to jest coś."),
	})))

	matches := rule.MatchList([]*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("To jest grapefruit. Dobry grejpfrut."),
	})
	require.Equal(t, 1, len(matches))
	require.Equal(t, "grapefruit", matches[0].GetSuggestedReplacements()[0])

	// Single-sentence mixed variants (positions relative to that sentence).
	matches = rule.MatchList([]*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("To jest jego bluff. A może blef?"),
	})
	require.Equal(t, 1, len(matches))
	require.Equal(t, "bluff", matches[0].GetSuggestedReplacements()[0])
}

func TestWordCoherencyRule_CallIndependence(t *testing.T) {
	rule := NewWordCoherencyRule(nil)
	require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("To jest blef."),
	})))
	// Independent call — no memory of previous text
	require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("A to nie bluff."),
	})))
}

func TestWordCoherencyRule_MatchPosition(t *testing.T) {
	rule := NewWordCoherencyRule(nil)
	matches := rule.MatchList([]*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("To jest blef. Ale nie bluff."),
	})
	require.Equal(t, 1, len(matches))
	require.Equal(t, 22, matches[0].GetFromPos())
	require.Equal(t, 27, matches[0].GetToPos())
}

func TestWordCoherencyRule_FullForms(t *testing.T) {
	rule := NewWordCoherencyRule(nil)
	require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("To jest blef. Nie wierzysz? Nie widzisz blefu!"),
	})))
	require.Equal(t, 1, len(rule.MatchList([]*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("To jest blef. Nie wierzysz? Nie widzisz bluffu!"),
	})))
}
