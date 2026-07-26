package uk_test

// Twin of languagetool-language-modules/uk/src/test/java/org/languagetool/rules/uk/SimpleReplaceRuleTest.java
//
// Java setup:
//   private JLanguageTool lt = new JLanguageTool(Ukrainian.DEFAULT_VARIANT);
//   morfologikSpellerRule = (MorfologikUkrainianSpellerRule)lt.getLanguage().getDefaultSpellingRule();
//   rule = new SimpleReplaceRule(TestTools.getEnglishMessages(), morfologikSpellerRule, ukrainian);
//   rule.match(lt.getAnalyzedSentence(...));
//
// Go: real Ukrainian analysis — UkrainianTagger + official ukrainian.dict +
// UkrainianHybridDisambiguator (same wire as Soft/Renamed twins) + real
// MorfologikUkrainianSpellerRule (getSpeller1().getSuggestionsFromDefaultDicts).
// External package uk_test avoids import cycle with tagging/disambiguation/uk → rules/uk.

import (
	"strings"
	"testing"

	rulesuk "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/uk"
	"github.com/stretchr/testify/require"
)

// TestSimpleReplaceRule_Rule twins SimpleReplaceRuleTest.testRule — every active Java case.
func TestSimpleReplaceRule_Rule(t *testing.T) {
	lt := wireUKSoftLT(t)
	sp := rulesuk.NewMorfologikUkrainianSpellerRule()
	require.NotNil(t, sp.GetSpeller1(), "official uk_UA.dict required for :bad speller path")
	rule := rulesuk.NewSimpleReplaceRule(nil, sp)

	// correct sentences:
	require.Equal(t, 0, len(rule.Match(lt.GetAnalyzedSentence("Ці рядки повинні збігатися."))))

	// incorrect sentences:
	matches := rule.Match(lt.GetAnalyzedSentence("Ці рядки повинні співпадати"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, 2, len(matches[0].GetSuggestedReplacements()))
	require.Equal(t, []string{"збігатися", "сходитися"}, matches[0].GetSuggestedReplacements())
	require.False(t, strings.Contains(matches[0].GetMessage(), "просторічна форма"))

	matches = rule.Match(lt.GetAnalyzedSentence("Нападаючий"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"Нападник", "Нападальний", "Нападний"}, matches[0].GetSuggestedReplacements())

	// lemma "нападаючий" in replace.txt → dictionary path (title-case suggestions)
	matches = rule.Match(lt.GetAnalyzedSentence("Нападаючого"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"Нападник", "Нападальний", "Нападний"}, matches[0].GetSuggestedReplacements())

	// ignoreTagged: щедрота (:bad) matches dictionary; щедроти has good POS → ignore
	matches = rule.Match(lt.GetAnalyzedSentence("щедрота"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"щедрість", "гойність", "щедриня"}, matches[0].GetSuggestedReplacements())

	matches = rule.Match(lt.GetAnalyzedSentence("щедроти"))
	require.Equal(t, 0, len(matches))

	// subst tags → not good POS, not in replace / no :bad path match → 0
	matches = rule.Match(lt.GetAnalyzedSentence("200 чоловік."))
	require.Equal(t, 0, len(matches))

	matches = rule.Match(lt.GetAnalyzedSentence("Конрадом II і Генріхом III"))
	require.Equal(t, 0, len(matches))

	// :bad + speller → мікрорайону (space-containing suggestions filtered)
	// Java TODO: should not react at all — still asserts match + suggestion
	matches = rule.Match(lt.GetAnalyzedSentence("мікро-району"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"мікрорайону"}, matches[0].GetSuggestedReplacements())

	// subst + subst
	matches = rule.Match(lt.GetAnalyzedSentence("гребнем."))
	require.Equal(t, 0, len(matches))
}

// TestSimpleReplaceRule_Derivat twins SimpleReplaceRuleTest.testDerivat.
// Surface is also a key in replace.txt (dictionary path; same suggestions as Java).
func TestSimpleReplaceRule_Derivat(t *testing.T) {
	lt := wireUKSoftLT(t)
	sp := rulesuk.NewMorfologikUkrainianSpellerRule()
	rule := rulesuk.NewSimpleReplaceRule(nil, sp)

	matches := rule.Match(lt.GetAnalyzedSentence("перелиставши."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"перегорнувши", "прогортаючи"}, matches[0].GetSuggestedReplacements())
}

// TestSimpleReplaceRule_RulePartOfMultiword twins SimpleReplaceRuleTest.testRulePartOfMultiword.
func TestSimpleReplaceRule_RulePartOfMultiword(t *testing.T) {
	lt := wireUKSoftLT(t)
	sp := rulesuk.NewMorfologikUkrainianSpellerRule()
	rule := rulesuk.NewSimpleReplaceRule(nil, sp)

	matches := rule.Match(lt.GetAnalyzedSentence("на думку проводжаючих"))
	require.Equal(t, 1, len(matches))
}

// TestSimpleReplaceRule_Misspellings twins SimpleReplaceRuleTest.testMisspellings.
// Real speller: getSuggestionsFromDefaultDicts("ганделик") → [генделик].
func TestSimpleReplaceRule_Misspellings(t *testing.T) {
	lt := wireUKSoftLT(t)
	sp := rulesuk.NewMorfologikUkrainianSpellerRule()
	require.NotNil(t, sp.GetSpeller1())
	rule := rulesuk.NewSimpleReplaceRule(nil, sp)

	matches := rule.Match(lt.GetAnalyzedSentence("ганделик"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"генделик"}, matches[0].GetSuggestedReplacements())
	require.Equal(t, "Неправильно написане слово.", matches[0].GetMessage())
}

// TestSimpleReplaceRule_RuleByTag twins SimpleReplaceRuleTest.testRuleByTag.
// Java only asserts match count; real analysis tags adjp:actv…:bad
// (примкнувший may also hit dictionary — still one match, as Java asserts).
func TestSimpleReplaceRule_RuleByTag(t *testing.T) {
	lt := wireUKSoftLT(t)
	sp := rulesuk.NewMorfologikUkrainianSpellerRule()
	rule := rulesuk.NewSimpleReplaceRule(nil, sp)

	matches := rule.Match(lt.GetAnalyzedSentence("спороутворюючого"))
	require.Equal(t, 1, len(matches))
	// lemma not in replace.txt → adjp:actv:bad arm
	require.Contains(t, matches[0].GetMessage(), "Активні дієприкметники")

	matches = rule.Match(lt.GetAnalyzedSentence("примкнувшим"))
	require.Equal(t, 1, len(matches))
}

// TestSimpleReplaceRule_Meta twins Java id / description / MISC category.
func TestSimpleReplaceRule_Meta(t *testing.T) {
	rule := rulesuk.NewSimpleReplaceRule(nil, nil)
	require.Equal(t, "UK_SIMPLE_REPLACE", rule.GetID())
	require.Equal(t, "Пошук помилкових слів", rule.GetDescription())
	require.NotNil(t, rule.GetCategory())
}
