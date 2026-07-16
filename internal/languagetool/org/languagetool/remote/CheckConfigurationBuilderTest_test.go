package remote

// Twin of CheckConfigurationBuilderTest
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckConfigurationBuilder_Build(t *testing.T) {
	b := NewCheckConfigurationBuilder("en")
	c := b.Build()
	code, ok := c.GetLangCode()
	require.True(t, ok)
	require.Equal(t, "en", code)
}

func TestCheckConfigurationBuilder_InvalidConfig(t *testing.T) {
	require.Panics(t, func() { NewCheckConfigurationBuilder("") })
	b := NewAutoDetectCheckConfigurationBuilder()
	// enabledOnly without rules
	b.enabledOnly = true
	require.Panics(t, func() { b.Build() })
}

func TestCheckConfigurationBuilder_AutoDetect(t *testing.T) {
	b := NewAutoDetectCheckConfigurationBuilder()
	c := b.Build()
	require.True(t, c.IsGuessLanguage())
}
