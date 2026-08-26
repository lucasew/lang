package remote

// Twin of CheckConfigurationTest (Java king).
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Twin of CheckConfigurationTest.test
func TestCheckConfiguration_Test(t *testing.T) {
	c := NewCheckConfiguration("en", "", false, nil, false, nil, "", "", nil, "", "", "")
	code, ok := c.GetLangCode()
	require.True(t, ok)
	require.Equal(t, "en", code)
	require.Empty(t, c.GetMotherTongueLangCode())
	require.Empty(t, c.GetEnabledRuleIDs())
	require.False(t, c.IsEnabledOnly())
	require.False(t, c.IsGuessLanguage())
	require.Empty(t, c.GetDisabledRuleIDs())
}

// Twin of CheckConfigurationTest.testNull — null lists / invalid lang pair
func TestCheckConfiguration_Null(t *testing.T) {
	// Java: new CheckConfiguration(null, null, false, null, ...) → IAE
	// (lang null + !guessLanguage)
	require.Panics(t, func() {
		NewCheckConfiguration("", "", false, nil, false, nil, "", "", nil, "", "", "")
	})
	// nil receiver getters stay safe
	var c *CheckConfiguration
	_, ok := c.GetLangCode()
	require.False(t, ok)
	require.Empty(t, c.GetMotherTongueLangCode())
	require.False(t, c.IsGuessLanguage())
	require.Empty(t, c.GetEnabledRuleIDs())
	require.Empty(t, c.GetDisabledRuleIDs())
	require.Empty(t, c.GetMode())
}

func TestCheckConfiguration_Values(t *testing.T) {
	c := &CheckConfiguration{
		LangCode:       "en",
		EnabledRuleIDs: []string{"A"},
		Mode:           "ALL",
	}
	code, ok := c.GetLangCode()
	require.True(t, ok)
	require.Equal(t, "en", code)
	require.Equal(t, []string{"A"}, c.GetEnabledRuleIDs())
}
