package ru

// Twin of languagetool-language-modules/ru/src/test/java/org/languagetool/rules/ru/RussianUnpairedBracketsRuleTest.java
//
// Java setup:
//   rule = new RussianUnpairedBracketsRule(TestTools.getEnglishMessages(), Russian.getInstance());
//   lt   = new JLanguageTool(Russian.getInstance());
// assert: rule.match(Collections.singletonList(lt.getAnalyzedSentence(...)))
// AnalyzePlain is outcome-identical for this suite (symbol tokens only).

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func ruMessagesUBR() map[string]string {
	// MessagesBundle_en keys used by GenericUnpairedBracketsRule / category (Java TestTools.getEnglishMessages()).
	return map[string]string{
		"unpaired_brackets":      "Unpaired symbol: ''{0}'' seems to be missing",
		"desc_unpaired_brackets": "Unpaired braces, brackets, quotation marks and similar symbols",
		"category_punctuation":   "Punctuation",
	}
}

// Port of RussianUnpairedBracketsRuleTest.testRuleRussian
func TestRussianUnpairedBracketsRule_RuleRussian(t *testing.T) {
	rule := NewRussianUnpairedBracketsRule(ruMessagesUBR())
	matchN := func(s string) int {
		return len(rule.MatchList([]*languagetool.AnalyzedSentence{languagetool.AnalyzePlain(s)}))
	}
	// correct sentences:
	require.Equal(t, 0, matchN("(О жене и детях не беспокойся, я беру их на свои руки)."))
	require.Equal(t, 0, matchN("Позже выходит другая «южная поэма» «Бахчисарайский фонтан» (1824)."))
	require.Equal(t, 0, matchN("А \"б\" Д."))
	// numerals / enumeration labels (Java NUMERALS_RU) — currently required twin case
	require.Equal(t, 0, matchN("а), б), Д)..., ДД), аа) и 1а)"))
	// incorrect sentences:
	require.Equal(t, 1, matchN("В таком ключе был начат в мае 1823 в Кишинёве роман в стихах 'Евгений Онегин."))
}

// Java RussianUnpairedBracketsRule: RU_UNPAIRED_BRACKETS, RU symbols, NUMERALS_RU, example pair.
func TestRussianUnpairedBracketsRule_Metadata(t *testing.T) {
	rule := NewRussianUnpairedBracketsRule(ruMessagesUBR())
	require.Equal(t, "RU_UNPAIRED_BRACKETS", rule.GetID())
	require.Equal(t, []string{"(", "{", "„", "\"", "'", "“"}, rule.StartSymbols)
	require.Equal(t, []string{")", "}", "“", "\"", "'", "”"}, rule.EndSymbols)
	// Java NUMERALS_RU pattern string (twin) + MatchString == Matcher.matches() samples
	require.Equal(t, `(?i)\d{1,2}?[а-я]*|[а-я]|[А-Я]|[а-я][а-я]|[А-Я][А-Я]|(?i)\d{1,2}?[a-z']*|M*(D?C{0,3}|C[DM])(L?X{0,3}|X[LC])(V?I{0,3}|I[VX])$`, NumeralsRUJava)
	for _, tok := range []string{"а", "б", "Д", "ДД", "аа", "1а"} {
		require.True(t, NumeralsRU.MatchString(tok), "NUMERALS_RU must match enumeration token %q", tok)
	}
	// full-string only (Java matches()): multi-letter words / bare dots must not match
	require.False(t, NumeralsRU.MatchString("жене"))
	require.False(t, NumeralsRU.MatchString("..."))
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "PUNCTUATION", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSTypographical, rule.GetLocQualityIssueType())
	require.Equal(t, ruMessagesUBR()["desc_unpaired_brackets"], rule.GetDescription())
	require.Equal(t, -1, rule.MinToCheckParagraph())
	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "Самоотверженный поступок Оленина <marker>(</marker>подарок Лукашке коня вызывает лишь удивление и усиливает недоверие к нему станичников.", inc[0].GetExample())
	// Java Rule.addExamplePair uses first fixed <marker> span as correction
	require.Equal(t, []string{"("}, inc[0].GetCorrections())
	require.Contains(t, rule.GetCorrectExamples()[0].GetExample(), "коня)")
	// MatchList rebinds Rule so match.Rule is the RU type with RU id
	ms := rule.MatchList([]*languagetool.AnalyzedSentence{
		languagetool.AnalyzePlain("В таком ключе был начат в мае 1823 в Кишинёве роман в стихах 'Евгений Онегин."),
	})
	require.NotEmpty(t, ms)
	require.Equal(t, rule, ms[0].Rule)
}
