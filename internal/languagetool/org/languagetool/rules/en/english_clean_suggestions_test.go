package en

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnglishCleanSuggestions(t *testing.T) {
	// single-token always kept
	require.Equal(t, []string{"rebuild"}, EnglishCleanSuggestions([]string{"rebuild"}))
	// "re ..." split dropped (#2562)
	require.Equal(t, []string{"ok"}, EnglishCleanSuggestions([]string{"re build", "ok"}))
	require.Equal(t, []string{"ok"}, EnglishCleanSuggestions([]string{"non compliant", "ok"}))
	// case-sensitive "i " prefix
	require.Equal(t, []string{"ok"}, EnglishCleanSuggestions([]string{"i phone", "ok"}))
	// ends with " able"
	require.Equal(t, []string{"ok"}, EnglishCleanSuggestions([]string{"read able", "ok"}))
}
