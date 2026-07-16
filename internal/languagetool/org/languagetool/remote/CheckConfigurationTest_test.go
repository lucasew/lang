package remote

// Twin of CheckConfigurationTest
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckConfiguration_Null(t *testing.T) {
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
