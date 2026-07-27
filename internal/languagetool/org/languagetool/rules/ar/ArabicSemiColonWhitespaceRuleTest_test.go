package ar

// Twin of languagetool-language-modules/ar/src/test/java/org/languagetool/rules/ar/ArabicSemiColonWhitespaceRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// arabicSemiColonMessages mirrors English MessagesBundle keys used by CommaWhitespaceRule
// (Java TestTools.getEnglishMessages()). Same keys as arabicCommaMessages.
func arabicSemiColonMessages() map[string]string {
	return map[string]string{
		"no_space_after":            "Don't put a space after the opening parenthesis",
		"no_space_around_quotes":    "Don't put a space after a quotation mark",
		"missing_space_after_comma": "Put a space after the comma",
		"no_space_before":           "Don't put a space before the closing parenthesis",
		"space_after_comma":         "Don't put a space before the comma",
		"no_space_before_dot":       "Don't put a space before the end-of-sentence period",
		"desc_comma_whitespace":     "Spacing around commas and parentheses",
		"category_typography":       "Typography",
	}
}

// TestArabicSemiColonWhitespaceRule_Metadata ports Java ctor/id/comma/default-on + super category/examples.
func TestArabicSemiColonWhitespaceRule_Metadata(t *testing.T) {
	rule := NewArabicSemiColonWhitespaceRule(arabicSemiColonMessages())
	require.Equal(t, "ARABIC_SC_WHITESPACE", rule.GetID())
	require.Equal(t, "؛", rule.GetCommaCharacter())
	require.False(t, rule.IsDefaultOff(), "Java does not call setDefaultOff")
	require.Equal(t, "Spacing around commas and parentheses", rule.GetDescription())
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "TYPOGRAPHY", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSWhitespace, rule.GetLocQualityIssueType())

	// Java super(messages, Example.wrong(...), Example.fixed(...)) → addExamplePair
	// Quirk preserved: example markers use Arabic question mark "؟", not "؛".
	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "نعم<marker> ؟</marker>", inc[0].GetExample())
	// fixed marker span "؟" becomes the correction (Rule.addExamplePair)
	require.Equal(t, []string{"؟"}, inc[0].GetCorrections())
	cor := rule.GetCorrectExamples()
	require.Len(t, cor, 1)
	require.Equal(t, "نعم<marker>؟</marker>", cor[0].GetExample())
}

// TestArabicSemiColonWhitespaceRule_Rule ports ArabicSemiColonWhitespaceRuleTest.testRule.
func TestArabicSemiColonWhitespaceRule_Rule(t *testing.T) {
	// Java: rule = new ArabicSemiColonWhitespaceRule(TestTools.getEnglishMessages());
	//       lt = new JLanguageTool(Languages.getLanguageForShortCode("ar"));
	rule := NewArabicSemiColonWhitespaceRule(arabicSemiColonMessages())
	assertMatches := func(text string, n int) {
		t.Helper()
		// Java: assertEquals(expectedMatches, rule.match(lt.getAnalyzedSentence(text)).length);
		// arAnalyze is defined in ArabicCommaWhitespaceRuleTest_test.go (same package).
		require.Equal(t, n, len(rule.Match(arAnalyze(text))), "text=%q", text)
	}

	// correct
	assertMatches("This is a test sentence؛", 0)
	assertMatches("أهذه تجربة؛", 0)

	// errors:
	// Arabic semi colon
	assertMatches("أهذه تجربة ؛", 1)
}

// TestArabicSemiColonWhitespaceRule_MatchRuleIdentity ensures RuleMatch.rule is the subclass
// (Java new RuleMatch(this, ...) with this = ArabicSemiColonWhitespaceRule).
func TestArabicSemiColonWhitespaceRule_MatchRuleIdentity(t *testing.T) {
	rule := NewArabicSemiColonWhitespaceRule(arabicSemiColonMessages())
	matches := rule.Match(arAnalyze("أهذه تجربة ؛"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "ARABIC_SC_WHITESPACE",
		matches[0].GetRule().(interface{ GetID() string }).GetID())
}
