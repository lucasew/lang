package ar

// Twin of languagetool-language-modules/ar/src/test/java/org/languagetool/rules/ar/ArabicCommaWhitespaceRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// arabicCommaMessages mirrors English MessagesBundle keys used by CommaWhitespaceRule
// (Java TestTools.getEnglishMessages()).
func arabicCommaMessages() map[string]string {
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

// arAnalyze ports Java JLanguageTool(Languages.getLanguageForShortCode("ar")).getAnalyzedSentence.
// Word-token stream matches ArabicWordTokenizer for the cases in ArabicCommaWhitespaceRuleTest
// (probed equal for all assertMatches texts: Arabic comma splits, ASCII comma, NBSP).
func arAnalyze(text string) *languagetool.AnalyzedSentence {
	return languagetool.NewJLanguageTool("ar").GetAnalyzedSentence(text)
}

// TestArabicCommaWhitespaceRule_Metadata ports Java ctor/id/comma/default-on + super category/examples.
func TestArabicCommaWhitespaceRule_Metadata(t *testing.T) {
	rule := NewArabicCommaWhitespaceRule(arabicCommaMessages())
	require.Equal(t, "ARABIC_COMMA_PARENTHESIS_WHITESPACE", rule.GetID())
	require.Equal(t, "،", rule.GetCommaCharacter())
	require.False(t, rule.IsDefaultOff(), "Java does not call setDefaultOff (unlike FA)")
	require.Equal(t, "Spacing around commas and parentheses", rule.GetDescription())
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "TYPOGRAPHY", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSWhitespace, rule.GetLocQualityIssueType())

	// Java super(messages, Example.wrong(...), Example.fixed(...)) → addExamplePair
	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "نعم<marker> ،</marker> لقد نجحنا.", inc[0].GetExample())
	// fixed marker span "،" becomes the correction (Rule.addExamplePair)
	require.Equal(t, []string{"،"}, inc[0].GetCorrections())
	cor := rule.GetCorrectExamples()
	require.Len(t, cor, 1)
	require.Equal(t, "نعم<marker>،</marker> لقد نجحنا.", cor[0].GetExample())
}

// TestArabicCommaWhitespaceRule_Rule ports ArabicCommaWhitespaceRuleTest.testRule.
func TestArabicCommaWhitespaceRule_Rule(t *testing.T) {
	// Java: rule = new ArabicCommaWhitespaceRule(TestTools.getEnglishMessages());
	//       lt = new JLanguageTool(Languages.getLanguageForShortCode("ar"));
	rule := NewArabicCommaWhitespaceRule(arabicCommaMessages())
	assertMatches := func(text string, n int) {
		t.Helper()
		// Java: assertEquals(expectedMatches, rule.match(lt.getAnalyzedSentence(text)).length);
		require.Equal(t, n, len(rule.Match(arAnalyze(text))), "text=%q", text)
	}

	// correct
	assertMatches("هذه جملة تجريبية.", 0)
	assertMatches("هذه, هي, جملة التجربة.", 0)
	assertMatches("قل (كيت وكيت) تجربة!.", 0)
	assertMatches("تكلف €2,45.", 0)
	assertMatches("ثمنها 50,- يورو", 0)
	assertMatches("جملة مع علامات الحذف ...", 0)
	assertMatches("هذه صورة: .5 وهي صحيحة.", 0)
	assertMatches("هذه $1,000,000.", 0)
	assertMatches("هذه 1,5.", 0)
	assertMatches("هذا ,,فحص''.", 0)
	assertMatches("نفّذ ./validate.sh لفحص الملف.", 0)
	assertMatches("هذه,\u00A0حقا,\u00A0فراغ غير فاصل.", 0)

	// errors — arabic comma (Java assertMatches counts)
	assertMatches("هذه،جملة للتجربة.", 1)
	assertMatches("هذه ، جملة للتجربة.", 1)
	assertMatches("هذه ،تجربة جملة.", 2)
	// Leading Arabic comma: "،" + "هذه" → space-before + missing-space-after = 2 (Java)
	assertMatches("،هذه جملة للتجربة.", 2)
}

// TestArabicCommaWhitespaceRule_MatchRuleIdentity ensures RuleMatch.rule is the subclass
// (Java new RuleMatch(this, ...) with this = ArabicCommaWhitespaceRule).
func TestArabicCommaWhitespaceRule_MatchRuleIdentity(t *testing.T) {
	rule := NewArabicCommaWhitespaceRule(arabicCommaMessages())
	matches := rule.Match(arAnalyze("هذه،جملة للتجربة."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, "ARABIC_COMMA_PARENTHESIS_WHITESPACE",
		matches[0].GetRule().(interface{ GetID() string }).GetID())
}
