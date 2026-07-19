package ga

// Twin of languagetool-language-modules/ga/src/test/java/org/languagetool/rules/ga/DhaNoBeirtRuleTest.java
import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestDhaNoBeirtRule_Rule(t *testing.T) {
	rule := NewDhaNoBeirtRule(nil)
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Seo abairt bheag."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Tá beirt dheartháireacha agam."))))
	// Java tagger: dheartháireacha lemma deartháir ∈ people.txt — inject lemma (no de-lenit invent).
	require.Equal(t, 1, len(rule.Match(withLemma("Tá dhá dheartháireacha agam.", "dheartháireacha", "deartháir"))))
	// "ab" is listed in people.txt as surface.
	require.Equal(t, 2, len(rule.Match(languagetool.AnalyzePlain("Seo dhá ab déag"))))
	require.Equal(t, 2, len(rule.Match(withLemma("Tá dhá dheartháireacha níos aosta déag agam.", "dheartháireacha", "deartháir"))))
}

func TestDhaNoBeirtRule_FailClosedWithoutLemma(t *testing.T) {
	rule := NewDhaNoBeirtRule(nil)
	// Lenited plural not in people.txt as surface; without lemma fail closed.
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Tá dhá dheartháireacha agam."))))
}

func withLemma(text, surface, lemma string) *languagetool.AnalyzedSentence {
	return languagetool.AnalyzeWithTagger(text, func(tok string) []languagetool.TokenTag {
		if strings.EqualFold(tok, surface) {
			return []languagetool.TokenTag{{POS: "Noun", Lemma: lemma}}
		}
		return nil
	})
}
