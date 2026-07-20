package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFilterMatchesByCategories(t *testing.T) {
	// Java AvsAnRule → Categories.MISC; Morfologik speller → TYPOS.
	ms := []LocalMatch{
		{RuleID: "EN_A_VS_AN", Message: "a/an"},
		{RuleID: "MORFOLOGIK_RULE_EN_US", Message: "spell"},
	}
	out := FilterMatchesByCategories(ms, []string{"MISC"}, nil, false)
	require.Len(t, out, 1)
	require.Equal(t, "MORFOLOGIK_RULE_EN_US", out[0].RuleID)

	out = FilterMatchesByCategories(ms, nil, []string{"MISC"}, true)
	require.Len(t, out, 1)
	require.Equal(t, "EN_A_VS_AN", out[0].RuleID)

	// Java: --enablecategories without --enabledonly does not hide other categories
	// (only enableRuleCategory; no disable of the rest).
	out = FilterMatchesByCategories(ms, nil, []string{"TYPOS"}, false)
	require.Len(t, out, 2)

	// LocalMatch CategoryID wins over RuleMeta (fixture IDs, not invent soft packs).
	ms2 := []LocalMatch{
		{RuleID: "FIXTURE_STYLE_RULE", CategoryID: "STYLE", Message: "style"},
		{RuleID: "FIXTURE_CASING_RULE", CategoryID: "CASING", Message: "case"},
	}
	out = FilterMatchesByCategories(ms2, []string{"STYLE"}, nil, false)
	require.Len(t, out, 1)
	require.Equal(t, "FIXTURE_CASING_RULE", out[0].RuleID)
	// enable STYLE without enabledOnly → both matches kept
	out = FilterMatchesByCategories(ms2, nil, []string{"STYLE"}, false)
	require.Len(t, out, 2)
	// enable STYLE with enabledOnly → only STYLE
	out = FilterMatchesByCategories(ms2, nil, []string{"STYLE"}, true)
	require.Len(t, out, 1)
	require.Equal(t, "FIXTURE_STYLE_RULE", out[0].RuleID)
}
