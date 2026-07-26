package ru

// Twin of languagetool-language-modules/ru/src/test/java/org/languagetool/rules/ru/RussianSpecificCaseRuleTest.java
//
// Analysis: Java uses JLanguageTool(ru).getAnalyzedSentence. For this surface
// casing rule, AnalyzePlain token boundaries match Java for the test strings
// (same approach as accepted EnglishSpecificCaseRule twin).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestRussianSpecificCaseRule_Rule(t *testing.T) {
	// Java: new RussianSpecificCaseRule(TestTools.getEnglishMessages())
	rule := NewRussianSpecificCaseRule(nil)

	assertGood := func(s string) {
		t.Helper()
		require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain(s))), "good %q", s)
	}
	assertBad := func(s string) []*rules.RuleMatch {
		t.Helper()
		matches := rule.Match(languagetool.AnalyzePlain(s))
		require.Equal(t, 1, len(matches), "bad %q", s)
		return matches
	}

	// Java RussianSpecificCaseRuleTest.testRule
	assertGood("Рытый Банк")
	assertGood("Центральный банк РФ")
	assertBad("Рытый банк")
	assertBad("центральный банк РФ")

	matches1 := assertBad("I like air France.")
	require.Equal(t, 7, matches1[0].GetFromPos())
	require.Equal(t, 17, matches1[0].GetToPos())
	require.Equal(t, []string{"Air France"}, matches1[0].GetSuggestedReplacements())
	require.Equal(t, "Для специальных наименований используйте начальную заглавную букву.", matches1[0].GetMessage())
}

// Java RussianSpecificCaseRule ctor metadata: CASING, Misspelling, example pair, resource /ru/specific_case.txt.
func TestRussianSpecificCaseRule_Metadata(t *testing.T) {
	rule := NewRussianSpecificCaseRule(nil)
	require.Equal(t, "RU_SPECIFIC_CASE", rule.GetID())
	require.Equal(t, "Написание специальных наименований в верхнем или нижнем регистре", rule.GetDescription())
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "CASING", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSMisspelling, rule.GetLocQualityIssueType())
	// Java does not setUrl
	require.Equal(t, "", rule.GetURL())

	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "Река <marker>рытый банк</marker> находится в Прикаспийской низменности.", inc[0].GetExample())
	require.Equal(t, []string{"Рытый Банк"}, inc[0].GetCorrections())
	// Java fixed example omits trailing period
	require.Equal(t, "Река <marker>Рытый Банк</marker> находится в Прикаспийской низменности", rule.GetCorrectExamples()[0].GetExample())

	// Resource embedded path twin of Java getFromResourceDirAsLines("/ru/specific_case.txt")
	m, maxLen := loadSpecificCase()
	require.NotEmpty(t, m)
	require.GreaterOrEqual(t, maxLen, 1)
	// known phrases from official specific_case.txt
	require.Equal(t, "Рытый Банк", m["рытый банк"])
	require.Equal(t, "Air France", m["air france"])
	require.Equal(t, "Центральный банк РФ", m["центральный банк рф"])

	// Java message branches (exact strings)
	require.Equal(t, "Для специальных наименований используйте начальную заглавную букву.", rule.InitialCapitalMessage)
	require.Equal(t, "Для специальных наименований используйте предложенное написание заглавных и строчных букв.", rule.OtherCapitalizationMessage)
	// Java RuleMatch has no custom short message; do not invent Russian short
	require.Equal(t, "", rule.ShortMsg)
}
