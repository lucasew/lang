package language

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestUkrainianAdvancedTypography_Disabled(t *testing.T) {
	// Java isAdvancedTypographyEnabled = false → no smart quotes / ellipsis
	require.False(t, UkrainianIsAdvancedTypographyEnabled())
	// suggestion tags only
	require.Equal(t, "«ok»", UkrainianAdvancedTypography("<suggestion>ok</suggestion>"))
	// ellipsis not applied when disabled
	require.Equal(t, "A...", UkrainianAdvancedTypography("A..."))
}

func TestUkrainianTypographyConfig_EnabledQuotes(t *testing.T) {
	cfg := UkrainianTypographyConfig(true)
	require.Equal(t, "«привіт»", languagetool.ToAdvancedTypography(`"привіт"`, cfg))
	require.Equal(t, "A…", languagetool.ToAdvancedTypography("A...", cfg))
}
