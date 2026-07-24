package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
	"github.com/stretchr/testify/require"
)

func TestRegisterCoreGermanRules_PriorityForIdWired(t *testing.T) {
	lt := languagetool.NewJLanguageTool("de")
	// Minimal: only wire hooks without full Register (faster) then full register
	RegisterCoreGermanRules(lt)
	require.NotNil(t, lt.PriorityForId)
	require.NotNil(t, lt.FilterRuleMatches)
	require.Equal(t, 10, lt.PriorityForId("OLD_SPELLING_RULE"))
	require.Equal(t, -15, lt.PriorityForId("STYLE"))
}

func TestRegisterCoreGermanRules_SwissFilterWired(t *testing.T) {
	lt := languagetool.NewJLanguageTool("de-CH")
	RegisterCoreGermanRules(lt)
	require.NotNil(t, lt.FilterRuleMatches)
	// Swiss filter rewrites ß → ss
	out := lt.FilterRuleMatches([]languagetool.LocalMatch{
		{RuleID: "X", Suggestions: []string{"groß"}},
	})
	require.Equal(t, []string{"gross"}, out[0].Suggestions)
}

func TestApplyRulePriorities_ViaCheck(t *testing.T) {
	lt := languagetool.NewJLanguageTool("de")
	lt.PriorityForId = language.GermanPriorityForId
	lt.AddRuleChecker("OLD_SPELLING_RULE", func(s *languagetool.AnalyzedSentence) []languagetool.LocalMatch {
		return []languagetool.LocalMatch{{
			FromPos: 0, ToPos: 1, RuleID: "OLD_SPELLING_RULE", Message: "old",
		}}
	})
	ms := lt.Check("x")
	require.NotEmpty(t, ms)
	require.Equal(t, 10, ms[0].Priority)
}
