package language

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUkrainianLanguage(t *testing.T) {
	require.Equal(t, "uk", UkrainianLanguageDefault.GetShortCode())
	require.Equal(t, "Ukrainian", UkrainianLanguageDefault.GetName())
	require.Contains(t, UkrainianLanguageDefault.GetCountries(), "UA")
	require.True(t, UkrainianIgnoredChars.MatchString("\u0301"))
	require.Contains(t, UkrainianLanguageDefault.RuleFiles, "grammar-spelling.xml")
}
