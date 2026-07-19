package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestEnglishPlainEnglishRule(t *testing.T) {
	rule := NewEnglishPlainEnglishRule(nil)
	// Java example: fatal outcome → death
	matches := rule.Match(languagetool.AnalyzePlain("a fatal outcome occurred"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "death", matches[0].GetSuggestedReplacements()[0])
}

// Java EnglishPlainEnglishRule: PLAIN_ENGLISH, Style, example pair.
func TestEnglishPlainEnglishRule_Metadata(t *testing.T) {
	rule := NewEnglishPlainEnglishRule(nil)
	require.Equal(t, "EN_PLAIN_ENGLISH_REPLACE", rule.GetID())
	require.Equal(t, "PLAIN_ENGLISH", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSStyle, rule.GetLocQualityIssueType())
	require.Equal(t, []string{"death"}, rule.GetIncorrectExamples()[0].GetCorrections())
}
