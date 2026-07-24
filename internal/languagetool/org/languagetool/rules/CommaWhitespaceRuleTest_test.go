package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.rules.CommaWhitespaceRuleTest — full-strength asserts.

func commaRule() *CommaWhitespaceRule {
	return NewCommaWhitespaceRule(map[string]string{
		"no_space_after":            "Don't put a space after the opening parenthesis",
		"no_space_around_quotes":    "Don't put a space after a quotation mark",
		"missing_space_after_comma": "Put a space after the comma",
		"no_space_before":           "Don't put a space before the closing parenthesis",
		"space_after_comma":         "Don't put a space before the comma",
		"no_space_before_dot":       "Don't put a space before the end-of-sentence period",
		"desc_comma_whitespace":     "Spacing around commas and parentheses",
	})
}

func assertMatches(t *testing.T, rule *CommaWhitespaceRule, text string, expected int) {
	t.Helper()
	n := len(rule.Match(languagetool.AnalyzePlain(text)))
	require.Equal(t, expected, n, "text=%q", text)
}

func TestCommaWhitespaceRule_Rule(t *testing.T) {
	rule := commaRule()

	assertMatches(t, rule, "This is a test sentence.", 0)
	assertMatches(t, rule, "I work with the technology .Net and Azure.", 0)
	assertMatches(t, rule, "I work with the technology .NET and Azure.", 0)
	assertMatches(t, rule, "I use .MP3 or .WAV file suffix", 0)
	assertMatches(t, rule, "This, is, a test sentence.", 0)
	assertMatches(t, rule, "This (foo bar) is a test!.", 0)
	assertMatches(t, rule, "Das kostet €2,45.", 0)
	assertMatches(t, rule, "Das kostet 50,- Euro", 0)
	assertMatches(t, rule, "This is a sentence with ellipsis ...", 0)
	assertMatches(t, rule, "This is a figure: .5 and it's correct.", 0)
	assertMatches(t, rule, "This is $1,000,000.", 0)
	assertMatches(t, rule, "This is 1,5.", 0)
	assertMatches(t, rule, "This is a ,,test''.", 0)
	assertMatches(t, rule, "Run ./validate.sh to check the file.", 0)
	assertMatches(t, rule, "This is,\u00A0really,\u00A0non-breaking whitespace.", 0)
	assertMatches(t, rule, "In his book,\u0002 Einstein proved this to be true.", 0)
	assertMatches(t, rule, "- [ ] A checkbox at GitHub", 0)
	assertMatches(t, rule, "- [x] A checked checkbox at GitHub", 0)
	assertMatches(t, rule, "A sentence 'with' ten \"correct\" examples of ’using’ quotation “marks” at «once» in it.", 0)
	assertMatches(t, rule, "I'd recommend resaving the .DOC as a PDF file.", 0)
	assertMatches(t, rule, "I'd recommend resaving the .mp3 as a WAV file.", 0)
	assertMatches(t, rule, "I'd suggest buying the .org domain.", 0)
	assertMatches(t, rule, ". This isn't good.", 0)
	assertMatches(t, rule, "), this isn't good.", 0)
	assertMatches(t, rule, "Das sind .exe-Dateien", 0)
	assertMatches(t, rule, "I live in .Los Angeles", 1)
	// Soft hyphens: Java Demo/DE ignoredCharactersRegex + assertMatchesForText (expect 1).
	// Use AnalyzeWithTokenizerAndIgnore (replaceSoftHyphens twin), not global U+00AD delete.
	t.Run("softHyphen", func(t *testing.T) {
		// after ignore-clean, missing space after comma still flags
		matches := rule.Match(languagetool.AnalyzeWithTokenizerAndIgnore(
			"This,is a soft\u00ADhyphen free comma.", nil, languagetool.GermanIgnoredCharactersRegex))
		require.Equal(t, 1, len(matches))
		// Twin of CommaWhitespaceRuleTest soft-hyphen German samples (expect 1 each).
		matches = rule.Match(languagetool.AnalyzeWithTokenizerAndIgnore(
			"Die Vertriebsniederlassu\u00ADng der Versorgungstechnik..\u00AD.",
			nil, languagetool.GermanIgnoredCharactersRegex))
		require.Equal(t, 1, len(matches), "soft-hyphen German ..\\u00AD.")
		// Multi-sentence path: trailing newline — Match still on first analyzed sentence.
		// Java assertMatchesForText sums over analyzeText sentences.
		matches = rule.Match(languagetool.AnalyzeWithTokenizerAndIgnore(
			"Die Vertriebsniederlassu\u00ADng der Versorgungstechnik..\u00AD.\n",
			nil, languagetool.GermanIgnoredCharactersRegex))
		require.Equal(t, 1, len(matches), "soft-hyphen German with trailing newline")
	})

	assertMatches(t, rule, "This,is a test sentence.", 1)
	assertMatches(t, rule, "This , is a test sentence.", 1)
	assertMatches(t, rule, "This ,is a test sentence.", 2)
	assertMatches(t, rule, ",is a test sentence.", 2)
	assertMatches(t, rule, "This ( foo bar) is a test!.", 1)
	assertMatches(t, rule, "This (foo bar ) is a test!.", 1)
	assertMatches(t, rule, "This is a sentence with an orphaned full stop .", 1)
	assertMatches(t, rule, "This is a test with a OOo footnote\u0002, which is denoted by 0x2 in the text.", 0)
	assertMatches(t, rule, "A sentence ' with ' ten \" incorrect \" examples of ’ using ’ quotation “ marks ” at « once » in it.", 10)
	assertMatches(t, rule, "A sentence ' with' one examples of wrong quotations marks in it.", 1)
	assertMatches(t, rule, "A sentence 'with ' one examples of wrong quotations marks in it.", 1)

	matches := rule.Match(languagetool.AnalyzePlain("ABB (  z.B. )"))
	require.Equal(t, 2, len(matches))
	require.Equal(t, 4, matches[0].GetFromPos())
	require.Equal(t, 6, matches[0].GetToPos())
	require.Equal(t, 11, matches[1].GetFromPos())
	require.Equal(t, 13, matches[1].GetToPos())

	matches = rule.Match(languagetool.AnalyzePlain("This ,"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, ",", matches[0].GetSuggestedReplacements()[0])
	matches = rule.Match(languagetool.AnalyzePlain("This ,is a test sentence."))
	require.Equal(t, 2, len(matches))
	require.Equal(t, ", ", matches[0].GetSuggestedReplacements()[0])
	matches = rule.Match(languagetool.AnalyzePlain("This , is a test sentence."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, ",", matches[0].GetSuggestedReplacements()[0])

	matches = rule.Match(languagetool.AnalyzePlain("You \" fixed\" it."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "\" ", matches[0].GetSuggestedReplacements()[0])
	require.Equal(t, " \"", matches[0].GetSuggestedReplacements()[1])
	require.Equal(t, 3, matches[0].GetFromPos())
	require.Equal(t, 6, matches[0].GetToPos())
	matches = rule.Match(languagetool.AnalyzePlain("You \"fixed \" it."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "\" ", matches[0].GetSuggestedReplacements()[0])
	require.Equal(t, " \"", matches[0].GetSuggestedReplacements()[1])
	require.Equal(t, 10, matches[0].GetFromPos())
	require.Equal(t, 13, matches[0].GetToPos())

	assertMatches(t, rule, "Ellipsis . . . as suggested by The Chicago Manual of Style", 3)
	assertMatches(t, rule, "Ellipsis . . . . as suggested by The Chicago Manual of Style", 4)
}
