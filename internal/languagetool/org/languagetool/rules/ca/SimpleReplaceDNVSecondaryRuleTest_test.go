package ca

// Twin of SimpleReplaceDNVSecondaryRuleTest — lemma + synthesizer path (no surface invent).
import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSimpleReplaceDNVSecondaryRule_Rule(t *testing.T) {
	rule := NewSimpleReplaceDNVSecondaryRule(nil)
	// Minimal synthesizer for participle/adj forms (Java CatalanSynthesizer).
	rule.Synthesize = func(lemma, postag string) []string {
		if lemma == "disposar" && strings.Contains(postag, "P") && strings.Contains(postag, "M") {
			// rough: past participle masculine singular-ish
			if strings.Contains(postag, "S") || strings.HasPrefix(postag, "VMP") {
				return []string{"disposat"}
			}
		}
		if lemma == "disposar" && postag == "VMP00SM0" {
			return []string{"disposat"}
		}
		if lemma == "disposar" {
			return []string{"disposat"}
		}
		if lemma == "baluerna" {
			return []string{"baluerna"}
		}
		return nil
	}

	// correct: lemma is not dispondre (disposar / indisposar adj) — untagged fail-closed
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Estan dispostes, estan indisposts, dispost a tot."))))

	// incorrect: surface dispost with lemma dispondre
	matches := rule.Match(analyzeCALemma("S'ha dispost a fer-ho.", map[string]languagetool.TokenTag{
		"dispost": {POS: "VMP00SM0", Lemma: "dispondre"},
	}))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "disposat", matches[0].GetSuggestedReplacements()[0])

	// armatost lemma
	matches = rule.Match(analyzeCALemma("Un armatost enorme.", map[string]languagetool.TokenTag{
		"armatost": {POS: "NCMS000", Lemma: "armatost"},
	}))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "baluerna", matches[0].GetSuggestedReplacements()[0])
}

func TestSimpleReplaceDNVSecondaryRule_FailClosedWithoutLemma(t *testing.T) {
	rule := NewSimpleReplaceDNVSecondaryRule(nil)
	require.Empty(t, rule.Match(languagetool.AnalyzePlain("S'ha dispost a fer-ho.")))
}

// analyzeCALemma injects FreeLing-style lemma+POS for named surfaces.
func analyzeCALemma(text string, tags map[string]languagetool.TokenTag) *languagetool.AnalyzedSentence {
	return languagetool.AnalyzeWithTagger(text, func(tok string) []languagetool.TokenTag {
		if tg, ok := tags[tok]; ok {
			return []languagetool.TokenTag{tg}
		}
		if tg, ok := tags[strings.ToLower(tok)]; ok {
			return []languagetool.TokenTag{tg}
		}
		return nil
	})
}
