package km

// Twin of languagetool-language-modules/km/src/test/java/org/languagetool/rules/km/KhmerSpaceBeforeRuleTest.java
//
// Java setup:
//
//	final Khmer language = new Khmer();
//	KhmerSpaceBeforeRule rule = new KhmerSpaceBeforeRule(TestTools.getEnglishMessages(), language);
//	JLanguageTool lt = new JLanguageTool(language);
//	// rule.match(lt.getAnalyzedSentence(text))
//
// Khmer uses KhmerWordTokenizer (ZWSP U+200B is a delimiter). Tests must use real
// km analysis, not AnalyzePlain (core WordTokenizer).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// TestKhmerSpaceBeforeRule_SpaceBeforeRule twins Java
// KhmerSpaceBeforeRuleTest.testSpaceBeforeRule (all 6 cases).
func TestKhmerSpaceBeforeRule_SpaceBeforeRule(t *testing.T) {
	rule := NewKhmerSpaceBeforeRule(nil)
	// Java: correct sentences
	assertKhmerSpaceBeforeMatches(t, rule, "គាត់​បាន​ទៅ ដើម្បី​ទិញ​ម្ហូប។", 0)
	assertKhmerSpaceBeforeMatches(t, rule, "ខ្ញុំ និង​គាត់។", 0)
	assertKhmerSpaceBeforeMatches(t, rule, "គាត់​ចង់​បាន ពីព្រោះ​គាត់​អត់​មាន។", 0)
	// Java: incorrect sentences
	assertKhmerSpaceBeforeMatches(t, rule, "គាត់​បាន​ទៅ​ដើម្បី​ទិញ​ម្ហូប។", 1)
	assertKhmerSpaceBeforeMatches(t, rule, "ខ្ញុំ​និង​គាត់។", 1)
	assertKhmerSpaceBeforeMatches(t, rule, "គាត់​ចង់​បាន​ពីព្រោះ​គាត់​អត់​មាន។", 1)
}

// TestKhmerSpaceBeforeRule_MatchDetails checks RuleMatch message/short/suggestion/span
// for one positive Java case (AbstractSpaceBeforeRule.match fields; UTF-16 lengths).
func TestKhmerSpaceBeforeRule_MatchDetails(t *testing.T) {
	rule := NewKhmerSpaceBeforeRule(nil)
	// "ខ្ញុំ​និង​គាត់។" — ZWSP before និង; previous is U+200B (not " " / "(")
	lt := languagetool.NewJLanguageTool("km")
	ms := rule.Match(lt.GetAnalyzedSentence("ខ្ញុំ​និង​គាត់។"))
	require.Len(t, ms, 1)
	// និង starts after ខ្ញុំ (5 UTF-16 units) + ZWSP (1) → start 6
	require.Equal(t, 6, ms[0].GetFromPos())
	// "និង" is three BMP code units → UTF-16 length 3
	require.Equal(t, 9, ms[0].GetToPos())
	require.Equal(t, "Missing white space before conjunction", ms[0].GetMessage())
	require.Equal(t, "Missing white space", ms[0].GetShortMessage())
	require.Equal(t, []string{" និង"}, ms[0].GetSuggestedReplacements())
	require.Equal(t, rule.AbstractSpaceBeforeRule, ms[0].GetRule())
}

// TestKhmerSpaceBeforeRule_Metadata twins Java getId / getDescription / getShort /
// getSuggestion / CONJUNCTIONS / MISC category / not DefaultOff.
func TestKhmerSpaceBeforeRule_Metadata(t *testing.T) {
	rule := NewKhmerSpaceBeforeRule(nil)
	require.Equal(t, "KM_SPACE_BEFORE_CONJUNCTION", rule.GetID())
	require.Equal(t, "Checks for missing space before some conjunctions", rule.GetDescription())
	require.Equal(t, "Missing white space", rule.ShortMsg)
	require.Equal(t, "Missing white space before conjunction", rule.Suggestion)
	require.False(t, rule.IsDefaultOff(), "Java KM ctor does not call setDefaultOff()")
	require.NotNil(t, rule.GetCategory())
	require.True(t, rule.GetCategory().GetID().Equals(rules.CategoryMisc))
	// Java CONJUNCTIONS body (full-token match via AbstractSpaceBeforeRule)
	require.NotNil(t, rule.Conjunctions)
	require.Equal(t, `ដើម្បី|និង|ពីព្រោះ`, rule.Conjunctions.String())
	// Each alternation is a full-token conjunction under Matcher.matches()
	// (sentence-start: previous is SENT_START "" → 1 match)
	for _, conj := range []string{"ដើម្បី", "និង", "ពីព្រោះ"} {
		assertKhmerSpaceBeforeMatches(t, rule, conj, 1)
	}
}

// assertKhmerSpaceBeforeMatches ports Java:
//
//	assertEquals(expected, rule.match(lt.getAnalyzedSentence(text)).length)
func assertKhmerSpaceBeforeMatches(t *testing.T, rule *KhmerSpaceBeforeRule, text string, expected int) {
	t.Helper()
	// Java: new JLanguageTool(new Khmer()).getAnalyzedSentence(text)
	lt := languagetool.NewJLanguageTool("km")
	require.Equal(t, expected, len(rule.Match(lt.GetAnalyzedSentence(text))), "text=%q", text)
}
