package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPremiumOff(t *testing.T) {
	require.False(t, PremiumOff{}.IsPremiumRule("ANY"))
	require.False(t, IsPremiumVersion())
	require.True(t, IsPremiumStatusCheck(PremiumStatusCheckText))
	require.True(t, IsPremiumStatusCheck(PremiumStatusCheckText2))
	require.False(t, IsPremiumStatusCheck("hello"))
	SetTempNotPremiumRules([]string{"TEMP"})
	require.True(t, IsTempNotPremium("TEMP"))
	require.False(t, IsTempNotPremium("OTHER"))
}
