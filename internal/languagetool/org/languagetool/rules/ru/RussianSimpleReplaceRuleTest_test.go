package ru

// Twin of languagetool-language-modules/ru/src/test/java/org/languagetool/rules/ru/RussianSimpleReplaceRuleTest.java
//
// Analysis path: AnalyzePlain (WordTokenizer + SENT_START) is equivalent to Java
// JLanguageTool(ru).getAnalyzedSentence for this rule’s surface matching.
// RussianSimpleReplaceRule/AbstractSimpleReplaceRule2 only inspects token surfaces,
// whitespaceBefore, and positions — not POS tags or lemmas. Official replace.txt
// keys tokenize the same under WordTokenizer as RU (accepted sibling pattern:
// RussianDashRule / EN SimpleReplaceRule).

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

// Java RussianSimpleReplaceRuleTest.testRule full matrix.
func TestRussianSimpleReplaceRule_Rule(t *testing.T) {
	// Java: new RussianSimpleReplaceRule(TestTools.getMessages("ru"));
	//       JLanguageTool lt = new JLanguageTool(Russian.getInstance());
	rule := NewRussianSimpleReplaceRule(nil)

	// correct sentences:
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Рост кораллов тут самый быстрый,"))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Книга была порвана."))))

	// incorrect sentences:
	matches := rule.Match(languagetool.AnalyzePlain("Книга была порвата."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, 1, len(matches[0].GetSuggestedReplacements()))
	require.Equal(t, "порвана", matches[0].GetSuggestedReplacements()[0])
	// Java-visible createMatch short / message templates (no sub-rule ids)
	require.Equal(t, "Ошибка?", matches[0].GetShortMessage())
	require.Equal(t,
		"«порвата» — просторечие, исправление: <suggestion>порвана</suggestion>",
		matches[0].GetMessage())
	// Java does not call useSubRuleSpecificIds() → match rule is the parent rule, not SpecificIdRule
	_, isSub := matches[0].GetRule().(*rules.SpecificIdRule)
	require.False(t, isSub, "no sub-rule specific id expected")
	base, ok := matches[0].GetRule().(*rules.AbstractSimpleReplaceRule2)
	require.True(t, ok)
	require.Equal(t, RussianSimpleReplaceRuleID, base.GetID())
}

// Field twin of RussianSimpleReplaceRule constructor + overrides.
func TestRussianSimpleReplaceRule_Metadata(t *testing.T) {
	rule := NewRussianSimpleReplaceRule(nil)

	require.Equal(t, RussianSimpleReplaceRuleID, rule.GetID())
	require.Equal(t, "RU_SIMPLE_REPLACE", rule.GetID())
	require.Equal(t, "Поиск просторечий и ошибочных фраз", rule.GetDescription())
	require.Equal(t, "Ошибка?", rule.GetShort())
	require.Equal(t, "«$match» — просторечие, исправление: $suggestions", rule.GetMessage())
	// Java getLocale() → new Locale("ru")
	require.Equal(t, "ru", rule.GetLocaleCode())
	require.Equal(t, RussianSimpleReplaceFileNames, rule.GetFileNames())
	require.Equal(t, []string{"/ru/replace.txt"}, rule.GetFileNames())
	// Java never calls useSubRuleSpecificIds()
	require.False(t, rule.SubRuleSpecificIDs)
	require.Equal(t, rules.CaseInsensitive, rule.CaseSens)
	// Java does not override getSuggestionsSeparator → default ", "
	// (empty Go field maps to ", " in createMatch)
	require.Equal(t, "", rule.SuggestionsSeparator)

	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "MISC", rule.GetCategory().GetID().String())
	// Java: setCategory(new Category(new CategoryId("MISC"), "Общие правила"))
	require.Equal(t, "Общие правила", rule.GetCategory().GetName())
	require.Equal(t, rules.ITSMisspelling, rule.GetLocQualityIssueType())

	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t,
		"<marker>Экспрессо</marker> – крепкий кофе, приготовленный из хорошо обжаренных и тонко помолотых кофейных зёрен.",
		inc[0].GetExample())
	require.Equal(t, []string{"Эспрессо"}, inc[0].GetCorrections())
	corr := rule.GetCorrectExamples()
	require.Len(t, corr, 1)
	require.Equal(t,
		"<marker>Эспрессо</marker> – крепкий кофе, приготовленный из хорошо обжаренных и тонко помолотых кофейных зёрен.",
		corr[0].GetExample())
}
