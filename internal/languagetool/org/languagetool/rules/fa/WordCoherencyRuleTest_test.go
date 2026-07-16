package fa

// Twin of languagetool-language-modules/fa/src/test/java/org/languagetool/rules/fa/WordCoherencyRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestWordCoherencyRule_Rules(t *testing.T) {
	rule := NewWordCoherencyRule(nil)
	require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("این یک اتاق است."),
	})))
	// two variants across sentences in one list
	require.Equal(t, 1, len(rule.MatchList([]*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("این یک اتاق است."),
		languagetool.AnalyzePlain("این یک اطاق است."),
	})))
}
