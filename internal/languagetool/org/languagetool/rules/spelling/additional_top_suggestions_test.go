package spelling

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAdditionalTopSuggestions_LanguageTool(t *testing.T) {
	require.Equal(t, []string{"LanguageTool"}, AdditionalTopSuggestions(nil, "languagetool"))
	require.Equal(t, []string{"LanguageTool"}, AdditionalTopSuggestions(nil, "Languagetool"))
	require.Equal(t, []string{"LanguageTooler"}, AdditionalTopSuggestions(nil, "languagetooler"))
	// already present → no add
	require.Empty(t, AdditionalTopSuggestions([]string{"LanguageTool"}, "languagetool"))
	require.Empty(t, AdditionalTopSuggestions(nil, "other"))
}
