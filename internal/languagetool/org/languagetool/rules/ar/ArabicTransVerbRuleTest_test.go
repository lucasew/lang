package ar

// Twin of ArabicTransVerbRuleTest — lemma+POS path (no surface clitic invent).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestArabicTransVerbRule_Rule(t *testing.T) {
	rule := NewArabicTransVerbRule(nil)
	require.NotNil(t, rule)

	// Bare untagged: fail closed
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("أفاض"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("أفاضه الماء"))))

	// Attached transitive verb lemma أفاض with POS, next not prep في
	sent := languagetool.AnalyzeWithTagger("أفاضه الماء", func(tok string) []languagetool.TokenTag {
		if tok == "أفاضه" {
			return []languagetool.TokenTag{{POS: "V", Lemma: "أفاض"}}
		}
		if tok == "الماء" {
			return []languagetool.TokenTag{{POS: "N", Lemma: "ماء"}}
		}
		return nil
	})
	matches := rule.Match(sent)
	require.NotEmpty(t, matches)
	require.Contains(t, matches[0].GetSuggestedReplacements()[0], "في")

	// Correct: next lemma is the required preposition
	sentOK := languagetool.AnalyzeWithTagger("أفاض في الماء", func(tok string) []languagetool.TokenTag {
		if tok == "أفاض" {
			return []languagetool.TokenTag{{POS: "V", Lemma: "أفاض"}}
		}
		if tok == "في" {
			return []languagetool.TokenTag{{POS: "P", Lemma: "في"}}
		}
		return nil
	})
	require.Empty(t, rule.Match(sentOK))
}
