package fa

// Outcome twins for org.languagetool.rules.fa.PersianCommaWhitespaceRule.
// Java has no dedicated *Test.java for this type; behavior is CommaWhitespaceRule
// with getCommaCharacter() = "،" (U+060C), getId PERSIAN_COMMA_PARENTHESIS_WHITESPACE,
// and ctor setDefaultOff(). ArabicCommaWhitespaceRuleTest is the closest upstream
// fixture for Arabic-comma outcomes (same comma char; FA also uses PersianWordTokenizer
// which tokenizes on "،").

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/stretchr/testify/require"
)

func persianCommaMessages() map[string]string {
	// Mirror English MessagesBundle keys used by CommaWhitespaceRule (TestTools.getEnglishMessages).
	return map[string]string{
		"no_space_after":            "Don't put a space after the opening parenthesis",
		"no_space_around_quotes":    "Don't put a space after a quotation mark",
		"missing_space_after_comma": "Put a space after the comma",
		"no_space_before":           "Don't put a space before the closing parenthesis",
		"space_after_comma":         "Don't put a space before the comma",
		"no_space_before_dot":       "Don't put a space before the end-of-sentence period",
		"desc_comma_whitespace":     "Spacing around commas and parentheses",
	}
}

// faAnalyze ports Java JLanguageTool(Languages.getLanguageForShortCode("fa")).getAnalyzedSentence
// word-token path: PersianWordTokenizer, no tagger.
func faAnalyze(text string) *languagetool.AnalyzedSentence {
	return languagetool.AnalyzeWithTokenizer(text, tokenizers.NewPersianWordTokenizer())
}

func TestPersianCommaWhitespaceRule_Metadata(t *testing.T) {
	rule := NewPersianCommaWhitespaceRule(persianCommaMessages())
	require.Equal(t, "PERSIAN_COMMA_PARENTHESIS_WHITESPACE", rule.GetID())
	require.Equal(t, "،", rule.GetCommaCharacter())
	require.True(t, rule.IsDefaultOff(), "Java ctor calls setDefaultOff()")
	require.Equal(t, "Spacing around commas and parentheses", rule.GetDescription())
}

func TestPersianCommaWhitespaceRule_Rule(t *testing.T) {
	rule := NewPersianCommaWhitespaceRule(persianCommaMessages())
	assertMatches := func(text string, n int) {
		t.Helper()
		require.Equal(t, n, len(rule.Match(faAnalyze(text))), "text=%q", text)
	}

	// correct — Arabic comma with proper spacing / non-comma cases
	assertMatches("بله،", 0)
	assertMatches("بله، تست", 0)
	assertMatches("این جمله آزمایشی.", 0)
	// ASCII comma is not the Persian comma character → no match from this rule
	assertMatches("این, تست.", 0)
	assertMatches("این,تست.", 0)
	assertMatches("قل (کیت و کیت) آزمایش!.", 0)
	assertMatches("تکست €2,45.", 0)
	assertMatches("قیمت 50,- یورو", 0)
	assertMatches("جمله با سه نقطه ...", 0)
	assertMatches("این تصویر: .5 درست است.", 0)
	assertMatches("این $1,000,000.", 0)
	assertMatches("این 1,5.", 0)
	assertMatches("این ,,بررسی''.", 0)
	assertMatches("اجرا کن ./validate.sh برای بررسی.", 0)
	assertMatches("این,\u00A0واقعا,\u00A0فاصله نشکن.", 0)

	// errors — Arabic comma (Java-visible CommaWhitespace outcomes with getCommaCharacter=،)
	// space before Arabic comma
	assertMatches("بله ،", 1)
	assertMatches("این ، جمله.", 1)
	// missing space after Arabic comma
	assertMatches("این،جمله", 1)
	// space before + missing space after
	assertMatches("این ،جمله", 2)
	// leading Arabic comma glued to word → space-before + missing-space-after
	assertMatches("،این جمله.", 2)

	// parentheses / period still from base CommaWhitespaceRule
	assertMatches("قل ( کیت)", 1)
	assertMatches("قل (کیت )", 1)
	assertMatches("تکست .", 1)
}

func TestPersianCommaWhitespaceRule_SpanAndSuggestion(t *testing.T) {
	rule := NewPersianCommaWhitespaceRule(persianCommaMessages())

	// Space before Arabic comma: "بله ،" → suggestion "،" (and "، " if no whitespace follows)
	matches := rule.Match(faAnalyze("بله ،"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, 3, matches[0].GetFromPos())
	require.Equal(t, 5, matches[0].GetToPos())
	require.Equal(t, []string{"،"}, matches[0].GetSuggestedReplacements())
	require.Equal(t, "PERSIAN_COMMA_PARENTHESIS_WHITESPACE",
		matches[0].GetRule().(interface{ GetID() string }).GetID())

	// Missing space after: "این،جمله" → suggestion "، جمله"
	matches = rule.Match(faAnalyze("این،جمله"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, 3, matches[0].GetFromPos())
	require.Equal(t, 8, matches[0].GetToPos())
	require.Equal(t, []string{"، جمله"}, matches[0].GetSuggestedReplacements())

	// Both errors: "بله ،تست"
	matches = rule.Match(faAnalyze("بله ،تست"))
	require.Equal(t, 2, len(matches))
	require.Equal(t, []string{"، "}, matches[0].GetSuggestedReplacements())
	require.Equal(t, []string{"، تست"}, matches[1].GetSuggestedReplacements())
}

func TestPersianCommaWhitespaceRule_RegisterDefaultOff(t *testing.T) {
	// Java Rule.setDefaultOff → JLanguageTool treats rule as default-off at registration.
	lt := languagetool.NewJLanguageTool("fa")
	RegisterCorePersianRules(lt)
	require.Contains(t, lt.GetDefaultOffRuleIDs(), "PERSIAN_COMMA_PARENTHESIS_WHITESPACE")
}
