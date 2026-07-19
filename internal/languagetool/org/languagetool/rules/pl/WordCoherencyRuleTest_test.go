package pl

// Twin of languagetool-language-modules/pl/src/test/java/org/languagetool/rules/pl/WordCoherencyRuleTest.java
// Production: file pairs only (no invent case suffixes). Inflected forms via lemmas.
import (
	"strings"
	"testing"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// twinCoherencyLemmas: surface → lemma for twin tests (Java morph / file bases).
var twinCoherencyLemmas = map[string]string{
	"grejpfrut": "grejpfrut", "grapefruit": "grapefruit",
	"blef": "blef", "bluff": "bluff",
	"blefy": "blef", "blefu": "blef",
	"bluffu": "bluff", "bluffy": "bluff",
}

func analyzePL(s string) *languagetool.AnalyzedSentence {
	return languagetool.AnalyzeWithTagger(s, plCoherencyTagWord)
}

func plCoherencyTagWord(tok string) []languagetool.TokenTag {
	key := strings.ToLower(tok)
	key = strings.TrimFunc(key, func(r rune) bool { return !unicode.IsLetter(r) })
	if lem, ok := twinCoherencyLemmas[key]; ok {
		return []languagetool.TokenTag{{Lemma: lem}}
	}
	return nil
}

func TestWordCoherencyRule_Rule(t *testing.T) {
	rule := NewWordCoherencyRule(nil)

	require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{
		analyzePL("To jest grejpfrut. Dobry grejpfrut."),
	})))
	require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{
		analyzePL("Lubię Twoje blefy. Blef to jest coś."),
	})))

	matches := rule.MatchList([]*languagetool.AnalyzedSentence{
		analyzePL("To jest grapefruit. Dobry grejpfrut."),
	})
	require.Equal(t, 1, len(matches))
	require.Equal(t, "grapefruit", matches[0].GetSuggestedReplacements()[0])

	// Single-sentence mixed variants (positions relative to that sentence).
	matches = rule.MatchList([]*languagetool.AnalyzedSentence{
		analyzePL("To jest jego bluff. A może blef?"),
	})
	require.Equal(t, 1, len(matches))
	require.Equal(t, "bluff", matches[0].GetSuggestedReplacements()[0])
}

func TestWordCoherencyRule_CallIndependence(t *testing.T) {
	rule := NewWordCoherencyRule(nil)
	require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{
		analyzePL("To jest blef."),
	})))
	// Independent call — no memory of previous text
	require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{
		analyzePL("A to nie bluff."),
	})))
}

func TestWordCoherencyRule_MatchPosition(t *testing.T) {
	rule := NewWordCoherencyRule(nil)
	matches := rule.MatchList([]*languagetool.AnalyzedSentence{
		analyzePL("To jest blef. Ale nie bluff."),
	})
	require.Equal(t, 1, len(matches))
	require.Equal(t, 22, matches[0].GetFromPos())
	require.Equal(t, 27, matches[0].GetToPos())
}

func TestWordCoherencyRule_FullForms(t *testing.T) {
	rule := NewWordCoherencyRule(nil)
	require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{
		analyzePL("To jest blef. Nie wierzysz? Nie widzisz blefu!"),
	})))
	require.Equal(t, 1, len(rule.MatchList([]*languagetool.AnalyzedSentence{
		analyzePL("To jest blef. Nie wierzysz? Nie widzisz bluffu!"),
	})))
}

func TestWordCoherencyRule_NoInventWithoutLemma(t *testing.T) {
	// Without lemmas, case forms not in coherency.txt must not invent-match.
	rule := NewWordCoherencyRule(nil)
	require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("To jest blef. Nie wierzysz? Nie widzisz bluffu!"),
	})))
}
