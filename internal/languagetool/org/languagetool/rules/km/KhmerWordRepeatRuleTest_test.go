package km

// Twin of languagetool-language-modules/km/src/test/java/org/languagetool/rules/km/KhmerWordRepeatRuleTest.java
//
// Java setup:
//
//	final Khmer language = new Khmer();
//	KhmerWordRepeatRule rule = new KhmerWordRepeatRule(TestTools.getEnglishMessages(), language);
//	JLanguageTool lt = new JLanguageTool(language);
//	// rule.match(lt.getAnalyzedSentence(text))
//
// Khmer uses KhmerWordTokenizer (ZWSP U+200B is a delimiter). Tests must use real
// km analysis via NewJLanguageTool("km").GetAnalyzedSentence, not AnalyzePlain.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// kmWordRepeatEnglishMessages ports TestTools.getEnglishMessages() keys used by
// KhmerWordRepeatRule (MessagesBundle_en.properties).
func kmWordRepeatEnglishMessages() map[string]string {
	return map[string]string{
		"desc_repetition":       "Word repetition (e.g. 'will will')",
		"desc_repetition_short": "Word repetition",
		"repetition":            "Possible typo: you repeated a word.",
		// category_misc used by Categories.MISC.getCategory(messages)
		"category_misc": "Miscellaneous",
	}
}

// TestKhmerWordRepeatRule_WordRepeatRule twins Java
// KhmerWordRepeatRuleTest.testWordRepeatRule (all 4 cases).
func TestKhmerWordRepeatRule_WordRepeatRule(t *testing.T) {
	// Java: new KhmerWordRepeatRule(TestTools.getEnglishMessages(), language)
	rule := NewKhmerWordRepeatRule(kmWordRepeatEnglishMessages())
	// Java: correct sentences
	assertKhmerWordRepeatMatches(t, rule, "នេះ​ហើយៗ​នោះ។", 0)
	assertKhmerWordRepeatMatches(t, rule, "គាត់​ហើយ ហើយ​ខ្ញុំ។", 0)
	// Java: incorrect sentences
	assertKhmerWordRepeatMatches(t, rule, "នេះ​ហើយ​ហើយ​នោះ។", 1)
	assertKhmerWordRepeatMatches(t, rule, "ខ្ញុំ​និង​និង​គាត់។", 1)
}

// TestKhmerWordRepeatRule_MatchDetails checks RuleMatch message/short/suggestions/span
// for one positive Java case (KhmerWordRepeatRule.match fields; UTF-16 lengths).
func TestKhmerWordRepeatRule_MatchDetails(t *testing.T) {
	msgs := kmWordRepeatEnglishMessages()
	rule := NewKhmerWordRepeatRule(msgs)
	// "នេះ​ហើយ​ហើយ​នោះ។" — ZWSP-separated repeat of ហើយ
	lt := languagetool.NewJLanguageTool("km")
	ms := rule.Match(lt.GetAnalyzedSentence("នេះ​ហើយ​ហើយ​នោះ។"))
	require.Len(t, ms, 1)
	// first ហើយ starts at 4 (នេះ=3 + ZWSP=1); second at 8; span end 8+3=11
	require.Equal(t, 4, ms[0].GetFromPos())
	require.Equal(t, 11, ms[0].GetToPos())
	require.Equal(t, msgs["repetition"], ms[0].GetMessage())
	require.Equal(t, msgs["desc_repetition_short"], ms[0].GetShortMessage())
	// Java replacement order exact
	require.Equal(t, []string{"ហើយ ហើយ", "ហើយ", "ហើយៗ"}, ms[0].GetSuggestedReplacements())
	require.Equal(t, rule, ms[0].GetRule())
}

// TestKhmerWordRepeatRule_Metadata twins Java getId / getDescription / MISC category.
func TestKhmerWordRepeatRule_Metadata(t *testing.T) {
	msgs := kmWordRepeatEnglishMessages()
	rule := NewKhmerWordRepeatRule(msgs)
	require.Equal(t, "KM_WORD_REPEAT_RULE", rule.GetID())
	require.Equal(t, msgs["desc_repetition"], rule.GetDescription())
	require.NotNil(t, rule.GetCategory())
	require.True(t, rule.GetCategory().GetID().Equals(rules.CategoryMisc))
	require.Equal(t, "Miscellaneous", rule.GetCategory().GetName())
}

// assertKhmerWordRepeatMatches ports Java:
//
//	assertEquals(expected, rule.match(lt.getAnalyzedSentence(text)).length)
func assertKhmerWordRepeatMatches(t *testing.T, rule *KhmerWordRepeatRule, text string, expected int) {
	t.Helper()
	// Java: new JLanguageTool(new Khmer()).getAnalyzedSentence(text)
	lt := languagetool.NewJLanguageTool("km")
	require.Equal(t, expected, len(rule.Match(lt.GetAnalyzedSentence(text))), "text=%q", text)
}
