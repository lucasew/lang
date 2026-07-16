package ar

// Twin of languagetool-language-modules/ar/src/test/java/org/languagetool/rules/ar/ArabicWordinessRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestArabicWordinessRule_Rule(t *testing.T) {
	rule := NewArabicWordinessRule(nil)
	require.Equal(t, 1, len(rule.Match(languagetool.AnalyzePlain("وأخيرا وليس آخرا"))))
}
