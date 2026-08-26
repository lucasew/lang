
package corepack_test

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/corepack"
	"github.com/stretchr/testify/require"
)

func TestRegister_FrenchFilterRuleMatchesWired(t *testing.T) {
	lt := languagetool.NewJLanguageTool("fr")
	corepack.Register(lt, "fr")
	require.NotNil(t, lt.FilterRuleMatches, "language init must wire French.filterRuleMatches")
	// Smoke: two adjacent AI_FR_GGEC merge via wired filter
	out := lt.FilterRuleMatches([]languagetool.LocalMatch{
		{FromPos: 0, ToPos: 2, RuleID: "AI_FR_GGEC_A", IssueType: "grammar", Suggestions: []string{"x"}},
		{FromPos: 2, ToPos: 4, RuleID: "AI_FR_GGEC_B", IssueType: "grammar", Suggestions: []string{"y"}},
	})
	require.Len(t, out, 1)
	require.Equal(t, "AI_FR_MERGED_MATCH", out[0].RuleID)
}

func TestRegister_FrenchPriorityForIdWired(t *testing.T) {
	lt := languagetool.NewJLanguageTool("fr")
	corepack.Register(lt, "fr")
	require.NotNil(t, lt.PriorityForId)
	require.Equal(t, 100, lt.PriorityForId("SA_CA_SE"))
	require.Equal(t, 500, lt.PriorityForId("FR_COMPOUNDS_X"))
	require.Equal(t, -101, lt.PriorityForId("AI_FR_HYDRA_LEO_X"))
}
