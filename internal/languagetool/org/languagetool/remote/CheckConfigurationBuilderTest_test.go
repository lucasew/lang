package remote

// Twin of CheckConfigurationBuilderTest (Java king).
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Twin of CheckConfigurationBuilderTest.test
func TestCheckConfigurationBuilder_Test(t *testing.T) {
	config1 := NewCheckConfigurationBuilder("xx").Build()
	code, ok := config1.GetLangCode()
	require.True(t, ok)
	require.Equal(t, "xx", code)
	require.Empty(t, config1.GetMotherTongueLangCode())
	require.Empty(t, config1.GetEnabledRuleIDs())
	require.False(t, config1.IsEnabledOnly())
	require.False(t, config1.IsGuessLanguage())

	config2 := NewAutoDetectCheckConfigurationBuilder().
		SetMotherTongueLangCode("mm").
		EnabledOnly().
		EnabledRuleIDs("RULE1", "RULE2").
		DisabledRuleIDs("RULE3", "RULE4").
		Build()
	_, ok = config2.GetLangCode()
	require.False(t, ok)
	require.Equal(t, "mm", config2.GetMotherTongueLangCode())
	require.Equal(t, []string{"RULE1", "RULE2"}, config2.GetEnabledRuleIDs())
	require.Equal(t, []string{"RULE3", "RULE4"}, config2.GetDisabledRuleIDs())
	require.True(t, config2.IsEnabledOnly())
	require.True(t, config2.IsGuessLanguage())
}

func TestCheckConfigurationBuilder_InvalidConfig(t *testing.T) {
	// Java: enabledOnly without enabled rules → IllegalStateException
	require.Panics(t, func() {
		NewCheckConfigurationBuilder("xx").EnabledOnly().Build()
	})
	require.Panics(t, func() { NewCheckConfigurationBuilder("") })
}

func TestCheckConfigurationBuilder_Build(t *testing.T) {
	b := NewCheckConfigurationBuilder("en")
	c := b.Build()
	code, ok := c.GetLangCode()
	require.True(t, ok)
	require.Equal(t, "en", code)
}

func TestCheckConfigurationBuilder_AutoDetect(t *testing.T) {
	b := NewAutoDetectCheckConfigurationBuilder()
	c := b.Build()
	require.True(t, c.IsGuessLanguage())
}
