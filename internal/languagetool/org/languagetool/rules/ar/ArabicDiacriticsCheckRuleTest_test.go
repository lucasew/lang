package ar

// Twin of ArabicDiacriticsCheckRuleTest
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of ArabicDiacriticsCheckRuleTest.testRule
func TestArabicDiacriticsCheckRule_Rule(t *testing.T) {
	rule := NewArabicDiacriticsRule(nil)
	matches := rule.Match(languagetool.AnalyzePlain("تجربة"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "تجرِبة", matches[0].GetSuggestedReplacements()[0])
}
