package ar

// Twin of languagetool-language-modules/ar/src/test/java/org/languagetool/rules/ar/ArabicWordCoherencyRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestArabicWordCoherencyRule_Rule(t *testing.T) {
	rule := NewArabicWordCoherencyRule(nil)
	// consistent text
	require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("كل شؤون العالم."),
	})))
	// mixed variants (pair from coherency.txt: شؤون;شئون)
	require.Equal(t, 1, len(rule.MatchList([]*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("كل شؤون العالم وكل شئون الناس."),
	})))
}
