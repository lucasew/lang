package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/UpperCaseNgramRuleTest.java
//
// Java:
//   Language lang = Languages.getLanguageForShortCode("de");
//   JLanguageTool lt = new JLanguageTool(lang);
//   Map: "5 Tagen"→100, "Sie tagen"→100;
//   FakeLanguageModel lm = new FakeLanguageModel(map);
//   UpperCaseNgramRule rule = new UpperCaseNgramRule(TestTools.getEnglishMessages(), lm, lang);
//   assertMatch(0/1, sentences...);

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/languagemodel"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ngrams"
	"github.com/stretchr/testify/require"
)

// fakeLanguageModelFromJavaMap ports FakeLanguageModel(map) as used by UpperCaseNgramRuleTest.
// Keys are space-joined ngrams (Java String.join(" ", tokens)); total = sum of all map values
// (FakeLanguageModel.getTotalTokenCount).
func fakeLanguageModelFromJavaMap(m map[string]int) ngrams.LanguageModel {
	lm := languagemodel.NewMapLanguageModel()
	var total int64
	for k, v := range m {
		c := int64(v)
		parts := strings.Split(k, " ")
		lm.Add(parts, c)
		total += c
	}
	lm.Total = total
	return lm
}

// TestUpperCaseNgramRule_Rule ports UpperCaseNgramRuleTest.testRule.
func TestUpperCaseNgramRule_Rule(t *testing.T) {
	// Java map + FakeLanguageModel
	mapCounts := map[string]int{
		"5 Tagen":   100,
		"Sie tagen": 100,
	}
	lm := fakeLanguageModelFromJavaMap(mapCounts)
	// Java: new UpperCaseNgramRule(TestTools.getEnglishMessages(), lm, lang)
	rule := NewUpperCaseNgramRule(nil, lm, "de")

	// Meta twin of Java ctor side-effects
	require.Equal(t, "DE_UPPER_CASE_NGRAM", rule.GetID())
	require.True(t, rule.IsDefaultTempOff())
	require.True(t, rule.IsDefaultOff())
	require.Equal(t, rules.ITSMisspelling, rule.GetLocQualityIssueType())
	require.NotNil(t, rule.GetCategory())
	require.Equal(t, rules.CatCasing.GetCategory(nil).GetID(), rule.GetCategory().GetID())
	require.NotEmpty(t, rule.GetIncorrectExamples())
	require.NotEmpty(t, rule.GetCorrectExamples())
	require.Equal(t, 50, upperCaseNgramThreshold)
	require.True(t, upperCaseNgramRelevantContains("tage"))
	require.True(t, upperCaseNgramRelevantContains("tagen"))
	require.True(t, upperCaseNgramRelevantContains("Tage"))
	require.True(t, upperCaseNgramRelevantContains("Tagen"))
	require.False(t, upperCaseNgramRelevantContains("other"))

	// Java: Languages.getLanguageForShortCode("de"); lt.getAnalyzedSentence(input)
	lt := languagetool.NewJLanguageTool("de")

	assertMatch := func(expectedMatches int, input string) {
		t.Helper()
		sent := lt.GetAnalyzedSentence(input)
		require.NotNil(t, sent, "GetAnalyzedSentence(%q)", input)
		matches := rule.Match(sent)
		require.Equal(t, expectedMatches, len(matches), "input %q got %#v", input, matches)
	}

	// Java assertMatch matrix (exactly four cases)
	assertMatch(0, "Nach 5 Tagen war es aus.")
	assertMatch(1, "Nach 5 tagen war es aus.")
	assertMatch(0, "Sie tagen im Hotel.")
	assertMatch(1, "Sie Tagen im Hotel.")

	// Outcome fidelity beyond length: suggestions + messages as Java Match builds them
	ms := rule.Match(lt.GetAnalyzedSentence("Nach 5 tagen war es aus."))
	require.Equal(t, 1, len(ms))
	require.Equal(t, []string{"Tagen"}, ms[0].GetSuggestedReplacements())
	require.Equal(t,
		"Meinten Sie das Nomen 'Tagen'? Nomen und Eigennamen werden großgeschrieben.",
		ms[0].GetMessage())

	ms = rule.Match(lt.GetAnalyzedSentence("Sie Tagen im Hotel."))
	require.Equal(t, 1, len(ms))
	require.Equal(t, []string{"tagen"}, ms[0].GetSuggestedReplacements())
	require.Equal(t,
		"Meinten Sie das Verb 'tagen'? Nur Nomen und Eigennamen werden großgeschrieben.",
		ms[0].GetMessage())
}
