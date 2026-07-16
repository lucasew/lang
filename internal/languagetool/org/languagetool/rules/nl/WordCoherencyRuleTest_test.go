package nl

// Twin of languagetool-language-modules/nl/src/test/java/org/languagetool/rules/nl/WordCoherencyRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestWordCoherencyRule_Rule(t *testing.T) {
	rule := NewWordCoherencyRule(nil)
	// "organogram, organigram" — two variants in one text
	sents := languagetool.AnalyzeTextLocal("organogram, organigram")
	require.Equal(t, 1, len(rule.MatchList(sents)))
}
