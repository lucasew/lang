package languagetool

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDynamicLanguage(t *testing.T) {
	d := NewDynamicLanguage("English", "en-US", "/dicts/en.dict")
	require.Equal(t, "en", d.GetShortCode())
	require.Equal(t, "en-US", d.GetShortCodeWithCountryAndVariant())
	require.Equal(t, "English", d.GetName())
	require.True(t, d.IsSpellcheckOnlyLanguage())
	require.Empty(t, d.GetRuleFileNames())
	require.Empty(t, d.GetPatternRules())
	require.Empty(t, d.GetCountries())
	require.Empty(t, d.GetMaintainers())
	require.Equal(t, filepath.Join("/dicts", "common_words.txt"), d.GetCommonWordsPath())
	// empty fields allowed (requireNonNull only)
	_ = NewDynamicLanguage("", "", "")
}
