package de

// Twin of UpperCaseNgramRuleTest — Java requires LanguageModel (no surface invent).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestUpperCaseNgramRule_WithoutLM_FailClosed(t *testing.T) {
	rule := NewUpperCaseNgramRule(nil)
	require.Empty(t, rule.Match(languagetool.AnalyzePlain("Nach 5 tagen war es aus.")))
	require.Empty(t, rule.Match(languagetool.AnalyzePlain("Sie Tagen im Hotel.")))
}

func TestUpperCaseNgramRule_WithProbability(t *testing.T) {
	rule := NewUpperCaseNgramRuleWithLM(nil, func(tri []string) float64 {
		if len(tri) != 3 {
			return 1e-20
		}
		// after "5", "Tagen" is common, "tagen" is not
		if tri[0] == "5" && tri[1] == "Tagen" {
			return 1.0
		}
		if tri[0] == "5" && tri[1] == "tagen" {
			return 0.001
		}
		return 0.01
	})
	ms := rule.Match(languagetool.AnalyzePlain("Nach 5 tagen war es aus."))
	require.Equal(t, 1, len(ms))
	require.Equal(t, "Tagen", ms[0].GetSuggestedReplacements()[0])
}
