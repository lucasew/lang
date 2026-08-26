package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.rules.CommaWhitespaceRuleTest — Java-visible outcomes only.
//
// Java setup:
//   rule = new CommaWhitespaceRule(TestTools.getEnglishMessages());
//   lt   = new JLanguageTool(TestTools.getDemoLanguage());
// Demo inherits Language.ignoredCharactersRegex = [\u00AD] (same pattern as German).

func commaRule() *CommaWhitespaceRule {
	// Messages mirror TestTools English ResourceBundle keys used by the rule.
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

// demoLT ports new JLanguageTool(TestTools.getDemoLanguage()) with Language base
// ignoredCharactersRegex = [\u00AD] (soft hyphen). GermanIgnoredCharactersRegex is
// the same pattern (Demo does not override getIgnoredCharactersRegex).
// SentenceTokenizer: Demo inherits Language.createDefaultSentenceTokenizer →
// SimpleSentenceTokenizer (segment-simple.srx), not full segment.srx for "xx".
func demoLT() *languagetool.JLanguageTool {
	lt := languagetool.NewJLanguageTool("xx")
	lt.IgnoredCharacters = languagetool.GermanIgnoredCharactersRegex
	lt.SentenceTokenizer = tokenizers.NewSimpleSentenceTokenizer()
	return lt
}

// assertMatches ports CommaWhitespaceRuleTest.assertMatches:
// rule.match(lt.getAnalyzedSentence(text)).length
func assertMatches(t *testing.T, rule *CommaWhitespaceRule, lt *languagetool.JLanguageTool, text string, expected int) {
	t.Helper()
	n := len(rule.Match(lt.GetAnalyzedSentence(text)))
	require.Equal(t, expected, n, "text=%q", text)
}

// assertMatchesForText ports CommaWhitespaceRuleTest.assertMatchesForText:
// sum of rule.match over lt.analyzeText(text) sentences.
func assertMatchesForText(t *testing.T, rule *CommaWhitespaceRule, lt *languagetool.JLanguageTool, text string, expected int) {
	t.Helper()
	matchCount := 0
	for _, s := range lt.Analyze(text) {
		matchCount += len(rule.Match(s))
	}
	require.Equal(t, expected, matchCount, "text=%q", text)
}

func TestCommaWhitespaceRule_Rule(t *testing.T) {
	rule := commaRule()
	lt := demoLT()

	assertMatches(t, rule, lt, "This is a test sentence.", 0)
	assertMatches(t, rule, lt, "I work with the technology .Net and Azure.", 0)
	assertMatches(t, rule, lt, "I work with the technology .NET and Azure.", 0)
	assertMatches(t, rule, lt, "I use .MP3 or .WAV file suffix", 0)
	assertMatches(t, rule, lt, "This, is, a test sentence.", 0)
	assertMatches(t, rule, lt, "This (foo bar) is a test!.", 0)
	assertMatches(t, rule, lt, "Das kostet €2,45.", 0)
	assertMatches(t, rule, lt, "Das kostet 50,- Euro", 0)
	assertMatches(t, rule, lt, "This is a sentence with ellipsis ...", 0)
	assertMatches(t, rule, lt, "This is a figure: .5 and it's correct.", 0)
	assertMatches(t, rule, lt, "This is $1,000,000.", 0)
	assertMatches(t, rule, lt, "This is 1,5.", 0)
	assertMatches(t, rule, lt, "This is a ,,test''.", 0)
	assertMatches(t, rule, lt, "Run ./validate.sh to check the file.", 0)
	assertMatches(t, rule, lt, "This is,\u00A0really,\u00A0non-breaking whitespace.", 0)
	// OpenOffice field codes:
	assertMatches(t, rule, lt, "In his book,\u0002 Einstein proved this to be true.", 0)
	assertMatches(t, rule, lt, "- [ ] A checkbox at GitHub", 0)
	assertMatches(t, rule, lt, "- [x] A checked checkbox at GitHub", 0)
	assertMatches(t, rule, lt, "A sentence 'with' ten \"correct\" examples of ’using’ quotation “marks” at «once» in it.", 0)
	assertMatches(t, rule, lt, "I'd recommend resaving the .DOC as a PDF file.", 0)
	assertMatches(t, rule, lt, "I'd recommend resaving the .mp3 as a WAV file.", 0)
	assertMatches(t, rule, lt, "I'd suggest buying the .org domain.", 0)
	assertMatches(t, rule, lt, ". This isn't good.", 0) // strange, but complaining without having a suggestions also doesn't help much...
	assertMatches(t, rule, lt, "), this isn't good.", 0)
	assertMatches(t, rule, lt, "Das sind .exe-Dateien", 0) // German, but the exception might be useful for other languages
	assertMatches(t, rule, lt, "I live in .Los Angeles", 1)
	// Soft-hyphen samples: Demo Language ignore [\u00AD] via getAnalyzedSentence / analyzeText.
	assertMatchesForText(t, rule, lt, "Die Vertriebsniederlassu\u00ADng der Versorgungstechnik..\u00AD.", 1)
	assertMatchesForText(t, rule, lt, "Die Vertriebsniederlassu\u00ADng der Versorgungstechnik..\u00AD.\n", 1)

	// errors:
	assertMatches(t, rule, lt, "This,is a test sentence.", 1)
	assertMatches(t, rule, lt, "This , is a test sentence.", 1)
	assertMatches(t, rule, lt, "This ,is a test sentence.", 2)
	assertMatches(t, rule, lt, ",is a test sentence.", 2)
	assertMatches(t, rule, lt, "This ( foo bar) is a test!.", 1)
	assertMatches(t, rule, lt, "This (foo bar ) is a test!.", 1)
	assertMatches(t, rule, lt, "This is a sentence with an orphaned full stop .", 1)
	assertMatches(t, rule, lt, "This is a test with a OOo footnote\u0002, which is denoted by 0x2 in the text.", 0)
	assertMatches(t, rule, lt, "A sentence ' with ' ten \" incorrect \" examples of ’ using ’ quotation “ marks ” at « once » in it.", 10)
	assertMatches(t, rule, lt, "A sentence ' with' one examples of wrong quotations marks in it.", 1)
	assertMatches(t, rule, lt, "A sentence 'with ' one examples of wrong quotations marks in it.", 1)

	matches := rule.Match(lt.GetAnalyzedSentence("ABB (  z.B. )"))
	require.Equal(t, 2, len(matches))
	require.Equal(t, 4, matches[0].GetFromPos())
	require.Equal(t, 6, matches[0].GetToPos())
	require.Equal(t, 11, matches[1].GetFromPos())
	require.Equal(t, 13, matches[1].GetToPos())

	matches = rule.Match(lt.GetAnalyzedSentence("This ,"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, ",", matches[0].GetSuggestedReplacements()[0])
	matches = rule.Match(lt.GetAnalyzedSentence("This ,is a test sentence."))
	require.Equal(t, 2, len(matches))
	require.Equal(t, ", ", matches[0].GetSuggestedReplacements()[0])
	matches = rule.Match(lt.GetAnalyzedSentence("This , is a test sentence."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, ",", matches[0].GetSuggestedReplacements()[0])

	matches = rule.Match(lt.GetAnalyzedSentence("You \" fixed\" it."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "\" ", matches[0].GetSuggestedReplacements()[0])
	require.Equal(t, " \"", matches[0].GetSuggestedReplacements()[1])
	require.Equal(t, 3, matches[0].GetFromPos())
	require.Equal(t, 6, matches[0].GetToPos())
	matches = rule.Match(lt.GetAnalyzedSentence("You \"fixed \" it."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "\" ", matches[0].GetSuggestedReplacements()[0])
	require.Equal(t, " \"", matches[0].GetSuggestedReplacements()[1])
	require.Equal(t, 10, matches[0].GetFromPos())
	require.Equal(t, 13, matches[0].GetToPos())

	assertMatches(t, rule, lt, "Ellipsis . . . as suggested by The Chicago Manual of Style", 3)
	assertMatches(t, rule, lt, "Ellipsis . . . . as suggested by The Chicago Manual of Style", 4)
}
