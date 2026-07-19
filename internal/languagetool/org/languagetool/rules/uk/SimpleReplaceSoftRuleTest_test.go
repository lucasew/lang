package uk

// Twin of languagetool-language-modules/uk/src/test/java/org/languagetool/rules/uk/SimpleReplaceSoftRuleTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestSimpleReplaceSoftRule_Rule(t *testing.T) {
	rule := NewSimpleReplaceSoftRule(nil)

	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Ці рядки повинні збігатися."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("у Трускавці."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("завидна"))))

	matches := rule.Match(languagetool.AnalyzePlain("Цей брелок"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"дармовис"}, matches[0].GetSuggestedReplacements())

	matches = rule.Match(languagetool.AnalyzePlain("Не знайде спасіння."))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"рятування", "рятунок", "порятунок", "визволення"}, matches[0].GetSuggestedReplacements())
	require.Contains(t, matches[0].GetMessage(), "релігія")
}

func TestSimpleReplaceSoftRule_RuleForDerivats(t *testing.T) {
	rule := NewSimpleReplaceSoftRule(nil)
	matches := rule.Match(languagetool.AnalyzePlain("Підключивши"))
	require.Equal(t, 1, len(matches))
	require.Equal(t, []string{"Увімкнувши", "Під'єднавши", "Приєднавши"}, matches[0].GetSuggestedReplacements())
}

func TestSimpleReplaceSoftRule_MetaAndCleanToken(t *testing.T) {
	rule := NewSimpleReplaceSoftRule(nil)
	require.Equal(t, "UK_SIMPLE_REPLACE_SOFT", rule.GetID())
	require.Equal(t, rules.ITSStyle, rule.GetLocQualityIssueType())
	require.NotNil(t, rule.GetCategory())

	// clean-token exception for завидна (Java getCleanToken)
	require.Empty(t, rule.Match(languagetool.AnalyzePlain("Завидна")))
}

func TestSimpleReplaceSoftRule_LemmaLookup(t *testing.T) {
	// checkLemmas true: surface form may miss, lemma of tagged token hits wrong-words
	// Soft map keys are lowercased surfaces; lemma path uses cleanup(lemma)
	rule := NewSimpleReplaceSoftRule(nil)
	// "брелок" as lemma on different surface if present
	bre := "брелок"
	tok := atrLemma("брелока", &bre, "noun:inanim:m:v_rod")
	sent := languagetool.NewAnalyzedSentence([]*languagetool.AnalyzedTokenReadings{tok})
	matches := rule.Match(sent)
	require.NotEmpty(t, matches)
	require.Contains(t, matches[0].GetSuggestedReplacements(), "дармовис")
}
