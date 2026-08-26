package languagetool

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestToAdvancedTypographyDisabled(t *testing.T) {
	cfg := DefaultTypographyConfig()
	cfg.Enabled = false
	out := ToAdvancedTypography("see <suggestion>foo</suggestion>", cfg)
	require.Equal(t, "see “foo”", out)
}

func TestToAdvancedTypographyEnabled(t *testing.T) {
	cfg := DefaultTypographyConfig()
	cfg.Enabled = true
	out := ToAdvancedTypography(`He said "hi"...`, cfg)
	require.Contains(t, out, "…")
	require.NotContains(t, out, "...")
	// apostrophe
	out2 := ToAdvancedTypography("don't", cfg)
	require.Contains(t, out2, "’")
}

func TestEqualsConsiderVariants(t *testing.T) {
	require.True(t, EqualsConsiderVariantsIfSpecified("en", "en-US"))
	require.True(t, EqualsConsiderVariantsIfSpecified("en-US", "en"))
	require.False(t, EqualsConsiderVariantsIfSpecified("en-US", "en-GB"))
	require.True(t, EqualsConsiderVariantsIfSpecified("de", "de"))
}

func TestAdaptSuggestionsList(t *testing.T) {
	require.Equal(t, []string{"a", "b"}, AdaptSuggestionsList([]string{"a", "b"}, "x"))
}
