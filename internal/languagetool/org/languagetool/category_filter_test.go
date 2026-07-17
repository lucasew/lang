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
}
