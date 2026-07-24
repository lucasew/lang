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
	// Java tempNotPremiumRules is fixed empty → always false
	require.False(t, IsTempNotPremium("TEMP"))
	require.False(t, IsTempNotPremium("ANY"))

	// deprecated build-info delegates (null when git-premium.properties absent)
	off := PremiumOff{}
	require.Nil(t, off.GetBuildDate())
	require.Nil(t, off.GetShortGitId())
	require.Nil(t, off.GetVersion())
}
