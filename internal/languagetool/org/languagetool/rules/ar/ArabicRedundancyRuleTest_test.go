package ar

// Twin of languagetool-language-modules/ar/src/test/java/org/languagetool/rules/ar/ArabicRedundancyRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestArabicRedundancyRule_Rule(t *testing.T) {
	rule := NewArabicRedundancyRule(nil)
	matches := rule.Match(languagetool.AnalyzePlain("سوف لن"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "لن", matches[0].GetSuggestedReplacements()[0])
}
