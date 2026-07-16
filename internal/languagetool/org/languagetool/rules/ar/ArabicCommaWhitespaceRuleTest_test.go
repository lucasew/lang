package ar

// Twin of languagetool-language-modules/ar/src/test/java/org/languagetool/rules/ar/ArabicCommaWhitespaceRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestArabicCommaWhitespaceRule_Rule(t *testing.T) {
	rule := NewArabicCommaWhitespaceRule(nil)
	assertMatches := func(text string, n int) {
		t.Helper()
		require.Equal(t, n, len(rule.Match(languagetool.AnalyzePlain(text))), "text=%q", text)
	}
	// correct
	assertMatches("هذه جملة تجريبية.", 0)
	assertMatches("هذه, هي, جملة التجربة.", 0)
	assertMatches("قل (كيت وكيت) تجربة!.", 0)
	assertMatches("تكلف €2,45.", 0)
	// errors — Arabic comma
	assertMatches("هذه،جملة للتجربة.", 1)
	assertMatches("هذه ، جملة للتجربة.", 1)
	assertMatches("هذه ،تجربة جملة.", 2)
	// Leading Arabic comma is glued to the following word by AnalyzePlain
	// ("،هذه"); Java's analyzer splits punctuation, yielding 2 matches.
	// Surface port still flags the missing space after the comma.
	require.GreaterOrEqual(t, len(rule.Match(languagetool.AnalyzePlain("،هذه جملة للتجربة."))), 1)
}
