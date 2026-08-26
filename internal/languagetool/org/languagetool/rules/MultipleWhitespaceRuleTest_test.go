package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.rules.MultipleWhitespaceRuleTest — Java-visible outcomes only.
//
// Java setup:
//   lt   = new JLanguageTool(TestTools.getDemoLanguage());
//   rule = new MultipleWhitespaceRule(TestTools.getEnglishMessages(), demo);
//   incorrect cases use lt.check(text) with only this rule enabled.
// Go: demoLT() + rule.Match on Analyze(text) (TextLevelRule list path).

func engMessagesMWS() map[string]string {
	// MessagesBundle_en keys used by MultipleWhitespaceRule.
	return map[string]string{
		"whitespace_repetition":    "Possible typo: you repeated a whitespace",
		"desc_whitespacerepetition": "Whitespace repetition",
	}
}

func multipleWhitespaceRule() *MultipleWhitespaceRule {
	return NewMultipleWhitespaceRule(engMessagesMWS())
}

// matchMWS ports lt.check for a TextLevelRule with only MultipleWhitespace enabled:
// analyzeText (SRX) then rule.match(sentences).
func matchMWS(rule *MultipleWhitespaceRule, lt *languagetool.JLanguageTool, text string) []*RuleMatch {
	return rule.Match(lt.Analyze(text))
}

// assertGoodWS ports MultipleWhitespaceRuleTest.assertGood → lt.check size 0.
func assertGoodWS(t *testing.T, rule *MultipleWhitespaceRule, lt *languagetool.JLanguageTool, input string) {
	t.Helper()
	matches := matchMWS(rule, lt, input)
	require.Equal(t, 0, len(matches), "assertGood: %q got %d matches", input, len(matches))
}

func TestMultipleWhitespaceRule_Rule(t *testing.T) {
	rule := multipleWhitespaceRule()
	lt := demoLT() // new JLanguageTool(TestTools.getDemoLanguage())

	// correct sentences:
	assertGoodWS(t, rule, lt, "This is a test sentence.")
	assertGoodWS(t, rule, lt, "This\uFEFF is a test sentence.")
	assertGoodWS(t, rule, lt, "This\uFEFF\uFEFF is a test sentence.")
	assertGoodWS(t, rule, lt, "This \uFEFFis a test sentence.")
	assertGoodWS(t, rule, lt, "This\uFEFF\u2060 is a test sentence.")
	assertGoodWS(t, rule, lt, "This\uFEFF\u2060 is a test sentence.")
	assertGoodWS(t, rule, lt, "\uFEFF\uFEFFThis is a\n\u2060\ntest sentence...")
	assertGoodWS(t, rule, lt, "This is a test sentence...")
	assertGoodWS(t, rule, lt, "\n\tThis is a test sentence...")
	assertGoodWS(t, rule, lt, "Multiple tabs\t\tare okay")
	assertGoodWS(t, rule, lt, "\n This is a test sentence...")
	assertGoodWS(t, rule, lt, "\n    This is a test sentence...")
	// Java comments out leading multi-space mid-document paragraph-start case.

	// incorrect sentences (Java: lt.check → count + from/to):
	matches := matchMWS(rule, lt, "This  is a test sentence.")
	require.Equal(t, 1, len(matches))
	require.Equal(t, 4, matches[0].GetFromPos())
	require.Equal(t, 6, matches[0].GetToPos())

	matches = matchMWS(rule, lt, "\n   This  is a test sentence.")
	require.Equal(t, 1, len(matches))
	require.Equal(t, 8, matches[0].GetFromPos())
	require.Equal(t, 10, matches[0].GetToPos())

	matches = matchMWS(rule, lt, "This is a test   sentence.")
	require.Equal(t, 1, len(matches))
	require.Equal(t, 14, matches[0].GetFromPos())
	require.Equal(t, 17, matches[0].GetToPos())

	matches = matchMWS(rule, lt, "This is   a  test   sentence.")
	require.Equal(t, 3, len(matches))
	require.Equal(t, 7, matches[0].GetFromPos())
	require.Equal(t, 10, matches[0].GetToPos())
	require.Equal(t, 11, matches[1].GetFromPos())
	require.Equal(t, 13, matches[1].GetToPos())
	require.Equal(t, 17, matches[2].GetFromPos())
	require.Equal(t, 20, matches[2].GetToPos())

	matches = matchMWS(rule, lt, "\t\t\t    \t\t\t\t  ")
	require.Equal(t, 2, len(matches))

	// with non-breakable spaces
	matches = matchMWS(rule, lt, "This \u00A0is a test sentence.")
	require.Equal(t, 1, len(matches))
	require.Equal(t, 4, matches[0].GetFromPos())
	require.Equal(t, 6, matches[0].GetToPos())
}
