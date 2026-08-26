package ar

// Outcome twins for org.languagetool.rules.ar.ArabicWordRepeatRule.
// Java has no dedicated ArabicWordRepeatRuleTest.java; phrases come from Java ignore()
// comments + example pair. Twin pattern matches IT ItalianWordRepeatRuleTest (real analysis).
//
// Java setup equivalent:
//
//	Arabic lang = new Arabic();
//	ArabicWordRepeatRule rule = new ArabicWordRepeatRule(TestTools.getEnglishMessages());
//	JLanguageTool lt = new JLanguageTool(lang);
//	// rule.match(lt.getAnalyzedSentence(text))
//
// Tests must use real AR analysis via NewJLanguageTool("ar").GetAnalyzedSentence / arAnalyze,
// not AnalyzePlain.
import (
	"testing"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// arabicWordRepeatEnglishMessages ports TestTools.getEnglishMessages() keys used by
// ArabicWordRepeatRule / WordRepeatRule (MessagesBundle_en.properties).
func arabicWordRepeatEnglishMessages() map[string]string {
	return map[string]string{
		"desc_repetition":       "Word repetition (e.g. 'will will')",
		"desc_repetition_short": "Word repetition",
		"repetition":            "Possible typo: you repeated a word.",
		"category_misc":         "Miscellaneous",
	}
}

// assertArabicWordRepeatMatches ports Java:
//
//	assertThat(rule.match(lt.getAnalyzedSentence(text)).length, is(expected))
func assertArabicWordRepeatMatches(t *testing.T, rule *ArabicWordRepeatRule, lt *languagetool.JLanguageTool, text string, expected int) {
	t.Helper()
	require.Equal(t, expected, len(rule.Match(lt.GetAnalyzedSentence(text))), "text=%q", text)
}

// TestArabicWordRepeatRule_Rule: ignore arms + example error + real repeat.
func TestArabicWordRepeatRule_Rule(t *testing.T) {
	// Java: new ArabicWordRepeatRule(TestTools.getEnglishMessages())
	rule := NewArabicWordRepeatRule(arabicWordRepeatEnglishMessages())
	// Java: new JLanguageTool(lang)
	lt := languagetool.NewJLanguageTool("ar")

	// Java ignore arm comment: "نفذت التعليمات خطوة خطوة." → 0
	assertArabicWordRepeatMatches(t, rule, lt, "نفذت التعليمات خطوة خطوة.", 0)
	// رويدا arm (tokenized as adjacent word pair)
	assertArabicWordRepeatMatches(t, rule, lt, "رويدا رويدا", 0)
	// super.ignore name list
	assertArabicWordRepeatMatches(t, rule, lt, "Duran Duran", 0)
	// Java example pair wrong sentence without markers → 1
	assertArabicWordRepeatMatches(t, rule, lt, "هذا فقط فقط مثال.", 1)
}

// ignoreAtAR mirrors Java-style:
//
//	tokens = lt.getAnalyzedSentence(input).getTokensWithoutWhitespace()
//	return rule.ignore(tokens, position)
func ignoreAtAR(t *testing.T, rule *ArabicWordRepeatRule, lt *languagetool.JLanguageTool, input string, position int) bool {
	t.Helper()
	tokens := lt.GetAnalyzedSentence(input).GetTokensWithoutWhitespace()
	return rule.Ignore(tokens, position)
}

// ignore() AR arms + super.ignore name list after ExtraIgnore returns false.
// Positions are the second token of the repeated pair (tokensWithoutWhitespace).
func TestArabicWordRepeatRule_Ignore(t *testing.T) {
	rule := NewArabicWordRepeatRule(arabicWordRepeatEnglishMessages())
	lt := languagetool.NewJLanguageTool("ar")

	// wordRepetitionOf exact equals at the second of the repeated pair
	// SENT_START=0: نفذت=1 التعليمات=2 خطوة=3 خطوة=4
	require.True(t, ignoreAtAR(t, rule, lt, "نفذت التعليمات خطوة خطوة.", 4))
	// رويدا=1 رويدا=2
	require.True(t, ignoreAtAR(t, rule, lt, "رويدا رويدا", 2))

	// Non-listed repeat is not ignored (هذا=1 فقط=2 فقط=3)
	require.False(t, ignoreAtAR(t, rule, lt, "هذا فقط فقط مثال.", 3))

	// super.ignore names (Java return super.ignore after AR checks)
	require.True(t, ignoreAtAR(t, rule, lt, "Duran Duran", 2))
	require.True(t, ignoreAtAR(t, rule, lt, "Phi Phi Islands", 2))
}

// wordRepetitionOf uses Token.equals (case-sensitive), not EqualFold.
// Arabic script has no case; Latin super names and the helper stay exact.
func TestArabicWordRepeatRule_WordRepetitionOfCaseSensitive(t *testing.T) {
	lt := languagetool.NewJLanguageTool("ar")

	// exact "خطوة" pair
	// SENT_START=0: نفذت=1 التعليمات=2 خطوة=3 خطوة=4
	tokens := lt.GetAnalyzedSentence("نفذت التعليمات خطوة خطوة.").GetTokensWithoutWhitespace()
	require.True(t, wordRepetitionOfAR("خطوة", tokens, 4))
	require.False(t, wordRepetitionOfAR("خطوة", tokens, 3))
	require.False(t, wordRepetitionOfAR("رويدا", tokens, 4))

	// exact "رويدا" pair
	tokensRo := lt.GetAnalyzedSentence("رويدا رويدا").GetTokensWithoutWhitespace()
	require.True(t, wordRepetitionOfAR("رويدا", tokensRo, 2))
	require.False(t, wordRepetitionOfAR("رويدا", tokensRo, 1))
	require.False(t, wordRepetitionOfAR("خطوة", tokensRo, 2))

	// Latin super list is case-sensitive via base wordRepetitionOf
	rule := NewArabicWordRepeatRule(nil)
	require.True(t, ignoreAtAR(t, rule, lt, "Duran Duran", 2))
	require.False(t, ignoreAtAR(t, rule, lt, "duran duran", 2))
	require.False(t, ignoreAtAR(t, rule, lt, "DURAN DURAN", 2))
}

// createRuleMatch outcomes: id, message, short message, span, suggestion.
func TestArabicWordRepeatRule_MatchSpanAndSuggestion(t *testing.T) {
	msgs := arabicWordRepeatEnglishMessages()
	rule := NewArabicWordRepeatRule(msgs)
	lt := languagetool.NewJLanguageTool("ar")
	text := "هذا فقط فقط مثال."
	sent := lt.GetAnalyzedSentence(text)
	matches := rule.Match(sent)
	require.Len(t, matches, 1)
	m := matches[0]
	require.Equal(t, "ARABIC_WORD_REPEAT_RULE", m.GetSpecificRuleId())
	require.Equal(t, msgs["repetition"], m.Message)
	require.Equal(t, msgs["desc_repetition_short"], m.ShortMessage)
	// Java: setSuggestedReplacement(prevToken) — surface of first "فقط"
	require.Equal(t, []string{"فقط"}, m.GetSuggestedReplacements())
	// Java: fromPos of first فقط through pos+prevToken.length() of second (UTF-16)
	require.Equal(t, "فقط فقط", utf16SubstrAR(text, m.FromPos, m.ToPos))
	require.Equal(t, rule.WordRepeatRule, m.GetRule())
}

// utf16SubstrAR slices text by Java UTF-16 code-unit offsets (token startPos/toPos).
func utf16SubstrAR(s string, from, to int) string {
	u := utf16.Encode([]rune(s))
	if from < 0 {
		from = 0
	}
	if to > len(u) {
		to = len(u)
	}
	if from >= to {
		return ""
	}
	return string(utf16.Decode(u[from:to]))
}

// Java getId + example pair + MISC category + Duplication + estimateContextForSureMatch + description key.
func TestArabicWordRepeatRule_Metadata(t *testing.T) {
	msgs := arabicWordRepeatEnglishMessages()
	rule := NewArabicWordRepeatRule(msgs)
	require.Equal(t, "ARABIC_WORD_REPEAT_RULE", rule.GetID())
	require.Equal(t, 1, rule.EstimateContextForSureMatch())
	require.Equal(t, msgs["desc_repetition"], rule.GetDescription())
	require.NotNil(t, rule.GetCategory())
	require.True(t, rule.GetCategory().GetID().Equals(rules.CategoryMisc))
	require.Equal(t, "Miscellaneous", rule.GetCategory().GetName())
	require.Equal(t, rules.ITSDuplication, rule.GetLocQualityIssueType())
	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "هذا <marker>فقط فقط</marker> مثال.", inc[0].GetExample())
	require.Equal(t, []string{"فقط"}, inc[0].GetCorrections())
	require.Equal(t, "هذا <marker>فقط</marker> مثال.", rule.GetCorrectExamples()[0].GetExample())
}
