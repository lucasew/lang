package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFilterMatchesByCategories(t *testing.T) {
	ms := []LocalMatch{
		{RuleID: "EN_A_VS_AN", Message: "a/an"},
		{RuleID: "MORFOLOGIK_RULE_EN_US", Message: "spell"},
	}
	out := FilterMatchesByCategories(ms, []string{"GRAMMAR"}, nil, false)
	require.Len(t, out, 1)
	require.Equal(t, "MORFOLOGIK_RULE_EN_US", out[0].RuleID)

	out = FilterMatchesByCategories(ms, nil, []string{"GRAMMAR"}, true)
	require.Len(t, out, 1)
	require.Equal(t, "EN_A_VS_AN", out[0].RuleID)

	// soft: --enablecategories without --enabledonly still restricts categories
	out = FilterMatchesByCategories(ms, nil, []string{"TYPOS"}, false)
	require.Len(t, out, 1)
	require.Equal(t, "MORFOLOGIK_RULE_EN_US", out[0].RuleID)

	// LocalMatch CategoryID wins over SoftRuleMeta
	ms2 := []LocalMatch{
		{RuleID: "EN_SOFT_X", CategoryID: "STYLE", Message: "style"},
		{RuleID: "EN_SOFT_Y", CategoryID: "CASING", Message: "case"},
	}
	out = FilterMatchesByCategories(ms2, []string{"STYLE"}, nil, false)
	require.Len(t, out, 1)
	require.Equal(t, "EN_SOFT_Y", out[0].RuleID)
	out = FilterMatchesByCategories(ms2, nil, []string{"STYLE"}, false)
	require.Len(t, out, 1)
	require.Equal(t, "EN_SOFT_X", out[0].RuleID)
}
