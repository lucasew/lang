package be

// Twin of languagetool-language-modules/be/src/test/java/org/languagetool/rules/be/BelarusianSpecificCaseRuleTest.java
//
// Analysis: Java uses JLanguageTool(be).getAnalyzedSentence. Go uses the same
// path (NewJLanguageTool("be").GetAnalyzedSentence) so token spans match Java.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestBelarusianSpecificCaseRule_Rule(t *testing.T) {
	// Java: new JLanguageTool(Languages.getLanguageForShortCode("be"))
	//       new BelarusianSpecificCaseRule(TestTools.getEnglishMessages())
	lt := languagetool.NewJLanguageTool("be")
	rule := NewBelarusianSpecificCaseRule(nil)

	assertGood := func(s string) {
		t.Helper()
		require.Equal(t, 0, len(rule.Match(lt.GetAnalyzedSentence(s))), "good %q", s)
	}
	assertBad := func(s string) []*rules.RuleMatch {
		t.Helper()
		matches := rule.Match(lt.GetAnalyzedSentence(s))
		require.Equal(t, 1, len(matches), "bad %q", s)
		return matches
	}

	// Java BelarusianSpecificCaseRuleTest.testRule
	assertGood("Беларуская Народная Рэспубліка")
	assertGood("Папа Рымскі")
	assertBad("дзяржаўны сцяг Рэспублікі Беларусь")
	assertBad("вярхоўны суд рэспублікі беларусь")

	matches1 := assertBad("Мне падабаецца air France.")
	require.Equal(t, 15, matches1[0].GetFromPos())
	require.Equal(t, 25, matches1[0].GetToPos())
	// Java: getSuggestedReplacements().toString() → "[Air France]"
	require.Equal(t, []string{"Air France"}, matches1[0].GetSuggestedReplacements())
	require.Equal(t, "Уласныя імёны і назвы пішуцца з вялікай літары.", matches1[0].GetMessage())
}

// Java BelarusianSpecificCaseRule ctor metadata: CASING, Misspelling, example pair,
// getPhrasesPath=/be/specific_case.txt, short message.
func TestBelarusianSpecificCaseRule_Metadata(t *testing.T) {
	rule := NewBelarusianSpecificCaseRule(nil)
	require.Equal(t, "BE_SPECIFIC_CASE", rule.GetID())
	require.Equal(t, "Напісанне спецыяльных найменняў у верхнім або ніжнім рэгістры", rule.GetDescription())
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, "CASING", rule.GetCategory().GetID().String())
	require.Equal(t, rules.ITSMisspelling, rule.GetLocQualityIssueType())
	// Java does not setUrl
	require.Equal(t, "", rule.GetURL())

	inc := rule.GetIncorrectExamples()
	require.Len(t, inc, 1)
	require.Equal(t, "<marker>Вялікая айчынная Вайна</marker> — гэта спецыфічны тэрмін савецкай гістарыяграфіі.", inc[0].GetExample())
	require.Equal(t, []string{"Вялікая Айчынная вайна"}, inc[0].GetCorrections())
	require.Equal(t, "<marker>Вялікая Айчынная вайна</marker> — гэта спецыфічны тэрмін савецкай гістарыяграфіі.", rule.GetCorrectExamples()[0].GetExample())

	// Resource twin of Java getFromResourceDirAsLines("/be/specific_case.txt")
	m, maxLen := loadSpecificCase()
	require.NotEmpty(t, m)
	require.GreaterOrEqual(t, maxLen, 2)
	require.Equal(t, "Беларуская Народная Рэспубліка", m["беларуская народная рэспубліка"])
	require.Equal(t, "Папа Рымскі", m["папа рымскі"])
	require.Equal(t, "Air France", m["air france"])
	require.Equal(t, "Дзяржаўны сцяг", m["дзяржаўны сцяг"])
	require.Equal(t, "Вярхоўны Суд Рэспублікі Беларусь", m["вярхоўны суд рэспублікі беларусь"])

	// Java message branches / short message (exact strings)
	require.Equal(t, "Уласныя імёны і назвы пішуцца з вялікай літары.", rule.InitialCapitalMessage)
	require.Equal(t, "Калі гэта уласнае імя або назва, выкарыстоўвайце прапанаванае напісанне.", rule.OtherCapitalizationMessage)
	require.Equal(t, "Proper noun", rule.ShortMsg)

	// RuleMatch short message from getShortMessage()
	lt := languagetool.NewJLanguageTool("be")
	ms := rule.Match(lt.GetAnalyzedSentence("air France"))
	require.Len(t, ms, 1)
	require.Equal(t, "Proper noun", ms[0].GetShortMessage())
}
