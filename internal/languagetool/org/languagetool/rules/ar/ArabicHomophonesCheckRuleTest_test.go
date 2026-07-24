package ar

// Twin of languagetool-language-modules/ar/src/test/java/org/languagetool/rules/ar/ArabicHomophonesCheckRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestArabicHomophonesCheckRule_Rule(t *testing.T) {
	rule := NewArabicHomophonesRule(nil)
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("ضن"))))
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("حاضر"))))
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("حض"))))
}
