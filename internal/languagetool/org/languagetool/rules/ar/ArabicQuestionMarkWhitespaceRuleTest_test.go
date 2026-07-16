package ar

// Twin of languagetool-language-modules/ar/src/test/java/org/languagetool/rules/ar/ArabicQuestionMarkWhitespaceRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestArabicQuestionMarkWhitespaceRule_Rule(t *testing.T) {
	rule := NewArabicQuestionMarkWhitespaceRule(nil)
	assertMatches := func(text string, n int) {
		t.Helper()
		require.Equal(t, n, len(rule.Match(languagetool.AnalyzePlain(text))), "text=%q", text)
	}
	assertMatches("أهذه تجربة؟", 0)
	assertMatches("This is a test sentence?", 0)
	assertMatches("أهذه تجربة?", 0)
	assertMatches("This is a test sentence؟", 0)
	assertMatches("أهذه تجربة ؟", 1)
}
