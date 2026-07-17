package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExpandSoftEnableRuleIDs(t *testing.T) {
	reg := []string{"EN_A_VS_AN", "EN_SOFT_OPT_PRIOR_TO", "EN_SOFT_OPT_GET_GO", "EN_SOFT_VERY_UNIQUE"}
	exp := ExpandSoftEnableRuleIDs(reg, []string{"SOFT_OPTIONAL", "EN_A_VS_AN"})
	require.Contains(t, exp, "EN_A_VS_AN")
	require.Contains(t, exp, "EN_SOFT_OPT_PRIOR_TO")
	require.Contains(t, exp, "EN_SOFT_OPT_GET_GO")
	require.NotContains(t, exp, "EN_SOFT_VERY_UNIQUE")
	require.Equal(t, []string{"EN_SOFT_OPT_PRIOR_TO"}, ExpandSoftEnableRuleIDs(reg, []string{"EN_SOFT_OPT_PRIOR_TO"}))
	require.Empty(t, ExpandSoftEnableRuleIDs(reg, nil))
}
