package corepack_test

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/corepack"
	"github.com/stretchr/testify/require"
)

func TestRegister_EnglishFilterRuleMatchesWired(t *testing.T) {
	lt := languagetool.NewJLanguageTool("en")
	corepack.Register(lt, "en")
	require.NotNil(t, lt.FilterRuleMatches)
	out := lt.FilterRuleMatches([]languagetool.LocalMatch{
		{RuleID: "EN_SIMPLE_REPLACE_PROGRAMME", Suggestions: []string{"program"}},
	})
	require.Equal(t, "locale-violation", out[0].IssueType)
}

func TestRegister_CatalanFiltersWired(t *testing.T) {
	lt := languagetool.NewJLanguageTool("ca")
	corepack.Register(lt, "ca")
	require.NotNil(t, lt.FilterRuleMatches)
	require.NotNil(t, lt.FilterRuleMatchesAfterOverlapping)
}

func TestRegister_EnglishPriorityForIdWired(t *testing.T) {
	lt := languagetool.NewJLanguageTool("en")
	corepack.Register(lt, "en")
	require.NotNil(t, lt.PriorityForId)
	require.Equal(t, 1, lt.PriorityForId("I_A"))
	require.Equal(t, -1, lt.PriorityForId("EN_A_VS_AN"))
	require.Equal(t, 2, lt.PriorityForId("EN_COMPOUNDS_X"))
	require.Equal(t, -49, lt.PriorityForId("EN_SIMPLE_REPLACE_PROGRAMME"))
}

func TestRegister_BritishEnglishPriorityForIdWired(t *testing.T) {
	lt := languagetool.NewJLanguageTool("en-GB")
	corepack.Register(lt, "en")
	require.NotNil(t, lt.PriorityForId)
	require.Equal(t, -20, lt.PriorityForId("OXFORD_SPELLING_ISATION_NOUNS"))
	require.Equal(t, 1, lt.PriorityForId("I_A"))
}

func TestRegister_CatalanPriorityForIdWired(t *testing.T) {
	lt := languagetool.NewJLanguageTool("ca")
	corepack.Register(lt, "ca")
	require.NotNil(t, lt.PriorityForId)
	require.Equal(t, 80, lt.PriorityForId("CONFUSIONS2"))
	require.Equal(t, -100, lt.PriorityForId("MORFOLOGIK_RULE_CA_ES"))
	require.Equal(t, 65, lt.PriorityForId("CA_SIMPLE_REPLACE_ANGLICISM_X"))
	require.Equal(t, -300, lt.PriorityForId("UPPERCASE_SENTENCE_START"))
}
