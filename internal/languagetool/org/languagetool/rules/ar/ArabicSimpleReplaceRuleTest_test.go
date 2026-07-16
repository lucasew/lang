package ar

// Twin of languagetool-language-modules/ar/src/test/java/org/languagetool/rules/ar/ArabicSimpleReplaceRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestArabicSimpleReplaceRule_Rule(t *testing.T) {
	rule := NewArabicSimpleReplaceRule(nil)

	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("عبد الله"))))

	matches := rule.Match(languagetool.AnalyzePlain("عبدالله"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "عبد الله", matches[0].GetSuggestedReplacements()[0])

	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("يافطة"))))
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("المائة"))))
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("الذى"))))
}
