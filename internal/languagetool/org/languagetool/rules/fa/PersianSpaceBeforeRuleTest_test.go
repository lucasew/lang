package fa

// Twin of languagetool-language-modules/fa/src/test/java/org/languagetool/rules/fa/PersianSpaceBeforeRuleTest.java
//
// Java setup:
//
//	rule = new PersianSpaceBeforeRule(TestTools.getEnglishMessages(), TestTools.getDemoLanguage());
//	lt = new JLanguageTool(TestTools.getDemoLanguage());
//
// Demo language uses default WordTokenizer (AnalyzePlain path). English messages are
// unused by this subclass (Persian hardcodes description/short/suggestion).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// TestPersianSpaceBeforeRule_Rules twins Java PersianSpaceBeforeRuleTest.testRules.
func TestPersianSpaceBeforeRule_Rules(t *testing.T) {
	rule := NewPersianSpaceBeforeRule(nil)
	// Java: assertMatches("به اینجا", 1);
	assertPersianSpaceBeforeMatches(t, rule, "به اینجا", 1)
	// Java: assertMatches("من به اینجا", 0);
	assertPersianSpaceBeforeMatches(t, rule, "من به اینجا", 0)
	// Java: assertMatches("(به اینجا", 0);
	assertPersianSpaceBeforeMatches(t, rule, "(به اینجا", 0)
}

// TestPersianSpaceBeforeRule_MatchDetails checks RuleMatch message/short/suggestion
// for the positive Java case (AbstractSpaceBeforeRule.match fields).
func TestPersianSpaceBeforeRule_MatchDetails(t *testing.T) {
	rule := NewPersianSpaceBeforeRule(nil)
	ms := rule.Match(languagetool.AnalyzePlain("به اینجا"))
	require.Len(t, ms, 1)
	require.Equal(t, 0, ms[0].GetFromPos())
	// "به" is two BMP code units → UTF-16 length 2
	require.Equal(t, 2, ms[0].GetToPos())
	require.Equal(t, "فاصلهٔ قبل از حرف ربط حذف شده‌است", ms[0].GetMessage())
	require.Equal(t, "فاصلهٔ حذف‌شده", ms[0].GetShortMessage())
	require.Equal(t, []string{" به"}, ms[0].GetSuggestedReplacements())
	require.Equal(t, rule.AbstractSpaceBeforeRule, ms[0].GetRule())
}

// TestPersianSpaceBeforeRule_Metadata twins Java getId / getDescription / getShort /
// getSuggestion / setDefaultOff / CONJUNCTIONS / MISC category.
func TestPersianSpaceBeforeRule_Metadata(t *testing.T) {
	rule := NewPersianSpaceBeforeRule(nil)
	require.Equal(t, "FA_SPACE_BEFORE_CONJUNCTION", rule.GetID())
	require.Equal(t, "بررسی‌کردن فاصله قبل از حرف ربط", rule.GetDescription())
	require.Equal(t, "فاصلهٔ حذف‌شده", rule.ShortMsg)
	require.Equal(t, "فاصلهٔ قبل از حرف ربط حذف شده‌است", rule.Suggestion)
	require.True(t, rule.IsDefaultOff(), "Java ctor calls setDefaultOff()")
	require.NotNil(t, rule.GetCategory())
	require.True(t, rule.GetCategory().GetID().Equals(rules.CategoryMisc))
	// Java CONJUNCTIONS body (full-token match via AbstractSpaceBeforeRule)
	require.NotNil(t, rule.Conjunctions)
	require.Equal(t, `و|به|با|تا|زیرا|چون|بنابراین|چونکه`, rule.Conjunctions.String())
	// Each alternation is a full-token conjunction under Matcher.matches()
	for _, conj := range []string{"و", "به", "با", "تا", "زیرا", "چون", "بنابراین", "چونکه"} {
		// Sentence-start conjunction: previous is SENT_START "" → 1 match
		assertPersianSpaceBeforeMatches(t, rule, conj, 1)
	}
	// Partial / non-full tokens must not match (matches() full-string)
	assertPersianSpaceBeforeMatches(t, rule, "چونک", 0)
	assertPersianSpaceBeforeMatches(t, rule, "xبه", 0)
}

// assertPersianSpaceBeforeMatches ports Java private assertMatches(text, expectedMatches).
func assertPersianSpaceBeforeMatches(t *testing.T, rule *PersianSpaceBeforeRule, text string, expected int) {
	t.Helper()
	// Java: rule.match(lt.getAnalyzedSentence(text)).length — Demo WordTokenizer
	require.Equal(t, expected, len(rule.Match(languagetool.AnalyzePlain(text))), "text=%q", text)
}
