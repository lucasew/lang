package rules

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of org.languagetool.rules.WordRepeatRuleTest — Java-visible outcomes only.
//
// Java setup:
//   Demo demoLanguage = new Demo();
//   JLanguageTool lt = new JLanguageTool(demoLanguage);
//   WordRepeatRule rule = new WordRepeatRule(TestTools.getEnglishMessages(), demoLanguage);
//   assertGood/Bad: rule.match(lt.getAnalyzedSentence(s)).length is 0 or 1

func wordRepeatRuleEnglish() *WordRepeatRule {
	// Messages mirror TestTools English ResourceBundle keys used by the rule.
	return NewWordRepeatRule(map[string]string{
		"repetition":            "Word repetition",
		"desc_repetition":       "Word repetition",
		"desc_repetition_short": "repetition",
		// Category name key used by Categories.MISC.getCategory(messages)
		"category_misc": "Miscellaneous",
	})
}

// assertWRGood ports WordRepeatRuleTest.assertGood:
// rule.match(lt.getAnalyzedSentence(s)).length == 0
func assertWRGood(t *testing.T, rule *WordRepeatRule, lt *languagetool.JLanguageTool, s string) {
	t.Helper()
	matches := rule.Match(lt.GetAnalyzedSentence(s))
	require.Equal(t, 0, len(matches), "assertGood %q", s)
}

// assertWRBad ports WordRepeatRuleTest.assertBad:
// rule.match(lt.getAnalyzedSentence(s)).length == 1
func assertWRBad(t *testing.T, rule *WordRepeatRule, lt *languagetool.JLanguageTool, s string) {
	t.Helper()
	matches := rule.Match(lt.GetAnalyzedSentence(s))
	require.Equal(t, 1, len(matches), "assertBad %q", s)
}

func TestWordRepeatRule_Test(t *testing.T) {
	// Java: new WordRepeatRule(TestTools.getEnglishMessages(), demoLanguage)
	// Language ctor arg is unused by core WordRepeatRule.match (kept for subclasses).
	rule := wordRepeatRuleEnglish()
	// demoLT ports new JLanguageTool(new Demo()) — see CommaWhitespaceRuleTest.
	lt := demoLT()

	assertWRGood(t, rule, lt, "A test")
	assertWRGood(t, rule, lt, "A test.")
	assertWRGood(t, rule, lt, "A test...")
	assertWRGood(t, rule, lt, "1 000 000 years")
	assertWRGood(t, rule, lt, "010 020 030")
	// Java: //👍💚🌲🌲  assertGood("\uD83D\uDC4D\uD83D\uDC9A\uD83C\uDF32\uD83C\uDF32");
	assertWRGood(t, rule, lt, "\U0001F44D\U0001F49A\U0001F332\U0001F332")

	assertWRBad(t, rule, lt, "A A test")
	assertWRBad(t, rule, lt, "A a test")
	assertWRBad(t, rule, lt, "This is is a test")
}

// Ctor fidelity: Java super.setCategory(Categories.MISC), setLocQualityIssueType(Duplication).
func TestWordRepeatRule_CategoryAndIssueType(t *testing.T) {
	r := NewWordRepeatRule(nil)
	require.NotNil(t, r.GetCategory())
	require.Equal(t, NewCategoryId("MISC"), r.GetCategory().GetID())
	require.Equal(t, ITSDuplication, r.GetLocQualityIssueType())
	require.Equal(t, "WORD_REPEAT_RULE", r.GetID())
	require.Equal(t, 1, r.EstimateContextForSureMatch())
}
