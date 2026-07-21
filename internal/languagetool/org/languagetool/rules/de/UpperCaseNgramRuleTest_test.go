package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/UpperCaseNgramRuleTest.java
// Java: FakeLanguageModel with map counts → BaseLanguageModel.getPseudoProbability.
// Without LM hook, Match is empty (fail-closed; no surface invent).
import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestUpperCaseNgramRule_WithoutLM_FailClosed(t *testing.T) {
	rule := NewUpperCaseNgramRule(nil)
	require.True(t, rule.DefaultTempOff)
	require.Equal(t, "DE_UPPER_CASE_NGRAM", rule.GetID())
	require.NotEmpty(t, rule.GetIncorrectExamples())
	require.Empty(t, rule.Match(languagetool.AnalyzePlain("Nach 5 tagen war es aus.")))
	require.Empty(t, rule.Match(languagetool.AnalyzePlain("Sie Tagen im Hotel.")))
}

// fakeNgramProb ports FakeLanguageModel + BaseLanguageModel.getPseudoProbability
// for the counts used in UpperCaseNgramRuleTest (map: "5 Tagen", "Sie tagen").
func fakeUpperCaseNgramProb(mapCounts map[string]int) func([]string) float64 {
	total := 0
	for _, v := range mapCounts {
		total += v
	}
	if total == 0 {
		total = 1
	}
	getCount := func(parts ...string) int {
		if len(parts) == 0 {
			return 0
		}
		// unigram and n-gram keys as space-joined (FakeLanguageModel.getCount)
		key := strings.Join(parts, " ")
		if c, ok := mapCounts[key]; ok {
			return c
		}
		return 0
	}
	return func(context []string) float64 {
		if len(context) == 0 {
			return 0
		}
		// BaseLanguageModel.getPseudoProbability chain rule
		firstWordCount := getCount(context[0])
		p := float64(firstWordCount+1) / float64(total+1)
		for i := 2; i <= len(context); i++ {
			sub := context[:i]
			phraseCount := getCount(sub...)
			thisP := float64(phraseCount+1) / float64(firstWordCount+1)
			p *= thisP
		}
		return p
	}
}

func TestUpperCaseNgramRule_Rule(t *testing.T) {
	// Java UpperCaseNgramRuleTest.testRule
	mapCounts := map[string]int{
		"5 Tagen":  100,
		"Sie tagen": 100,
	}
	rule := NewUpperCaseNgramRuleWithLM(nil, fakeUpperCaseNgramProb(mapCounts))

	assertMatch := func(expected int, input string) {
		t.Helper()
		ms := rule.Match(languagetool.AnalyzePlain(input))
		require.Equal(t, expected, len(ms), "input %q", input)
	}

	assertMatch(0, "Nach 5 Tagen war es aus.")
	assertMatch(1, "Nach 5 tagen war es aus.")
	assertMatch(0, "Sie tagen im Hotel.")
	assertMatch(1, "Sie Tagen im Hotel.")

	// suggestions / messages twin
	ms := rule.Match(languagetool.AnalyzePlain("Nach 5 tagen war es aus."))
	require.Equal(t, 1, len(ms))
	require.Equal(t, []string{"Tagen"}, ms[0].GetSuggestedReplacements())
	require.Contains(t, ms[0].GetMessage(), "Nomen")

	ms = rule.Match(languagetool.AnalyzePlain("Sie Tagen im Hotel."))
	require.Equal(t, 1, len(ms))
	require.Equal(t, []string{"tagen"}, ms[0].GetSuggestedReplacements())
	require.Contains(t, ms[0].GetMessage(), "Verb")
}

func TestUpperCaseNgramRule_WithProbability(t *testing.T) {
	// Direct ratio probe (still valid; Fake LM twin is TestUpperCaseNgramRule_Rule)
	rule := NewUpperCaseNgramRuleWithLM(nil, func(tri []string) float64 {
		if len(tri) != 3 {
			return 1e-20
		}
		if tri[0] == "5" && tri[1] == "Tagen" {
			return 1.0
		}
		if tri[0] == "5" && tri[1] == "tagen" {
			return 0.001
		}
		return 0.01
	})
	ms := rule.Match(languagetool.AnalyzePlain("Nach 5 tagen war es aus."))
	require.Equal(t, 1, len(ms))
	require.Equal(t, "Tagen", ms[0].GetSuggestedReplacements()[0])
}
