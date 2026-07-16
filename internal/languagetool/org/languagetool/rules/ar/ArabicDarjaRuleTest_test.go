package ar

// Twin of languagetool-language-modules/ar/src/test/java/org/languagetool/rules/ar/ArabicDarjaRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestArabicDarjaRule_Rule(t *testing.T) {
	rule := NewArabicDarjaRule(nil)
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("إن شاء"))))

	matches := rule.Match(languagetool.AnalyzePlain("طرشي"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "فلفل حلو", matches[0].GetSuggestedReplacements()[0])

	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("فايدة"))))
}
