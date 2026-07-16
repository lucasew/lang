package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.rules.MultipleWhitespaceRuleTest — full-strength asserts.

func engMessages() map[string]string {
	return map[string]string{
		"whitespace_repetition": "Possible typo: you repeated a whitespace",
		"desc_whitespacerepetition": "Whitespace repetition",
	}
}

func assertGoodWS(t *testing.T, rule *MultipleWhitespaceRule, input string) {
	t.Helper()
	sents := languagetool.AnalyzeSentences(input)
	matches := rule.Match(sents)
	require.Equal(t, 0, len(matches), "assertGood: %q got %d matches", input, len(matches))
}

func TestMultipleWhitespaceRule_Rule(t *testing.T) {
	rule := NewMultipleWhitespaceRule(engMessages())

	// correct sentences:
	assertGoodWS(t, rule, "This is a test sentence.")
	assertGoodWS(t, rule, "This\uFEFF is a test sentence.")
	assertGoodWS(t, rule, "This\uFEFF\uFEFF is a test sentence.")
	assertGoodWS(t, rule, "This \uFEFFis a test sentence.")
	assertGoodWS(t, rule, "This\uFEFF\u2060 is a test sentence.")
	assertGoodWS(t, rule, "This\uFEFF\u2060 is a test sentence.")
	assertGoodWS(t, rule, "\uFEFF\uFEFFThis is a\n\u2060\ntest sentence...")
	assertGoodWS(t, rule, "This is a test sentence...")
	assertGoodWS(t, rule, "\n\tThis is a test sentence...")
	assertGoodWS(t, rule, "Multiple tabs\t\tare okay")
	assertGoodWS(t, rule, "\n This is a test sentence...")
	assertGoodWS(t, rule, "\n    This is a test sentence...")

	// incorrect:
	matches := rule.Match(languagetool.AnalyzeSentences("This  is a test sentence."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, 4, matches[0].GetFromPos())
	require.Equal(t, 6, matches[0].GetToPos())

	matches = rule.Match(languagetool.AnalyzeSentences("\n   This  is a test sentence."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, 8, matches[0].GetFromPos())
	require.Equal(t, 10, matches[0].GetToPos())

	matches = rule.Match(languagetool.AnalyzeSentences("This is a test   sentence."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, 14, matches[0].GetFromPos())
	require.Equal(t, 17, matches[0].GetToPos())

	matches = rule.Match(languagetool.AnalyzeSentences("This is   a  test   sentence."))
	require.Equal(t, 3, len(matches))
	require.Equal(t, 7, matches[0].GetFromPos())
	require.Equal(t, 10, matches[0].GetToPos())
	require.Equal(t, 11, matches[1].GetFromPos())
	require.Equal(t, 13, matches[1].GetToPos())
	require.Equal(t, 17, matches[2].GetFromPos())
	require.Equal(t, 20, matches[2].GetToPos())

	matches = rule.Match(languagetool.AnalyzeSentences("\t\t\t    \t\t\t\t  "))
	require.Equal(t, 2, len(matches))

	matches = rule.Match(languagetool.AnalyzeSentences("This \u00A0is a test sentence."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, 4, matches[0].GetFromPos())
	require.Equal(t, 6, matches[0].GetToPos())
}
