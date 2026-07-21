package spelling

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSpellingCheckRule_PathGetters_Default(t *testing.T) {
	r := NewSpellingCheckRule("S", "d", "en-US")
	require.Equal(t, "en", r.ShortCode())
	require.Equal(t, "en/hunspell/ignore.txt", r.GetIgnoreFileName())
	require.Equal(t, "en/hunspell/spelling.txt", r.GetSpellingFileName())
	require.Equal(t, "en/hunspell/prohibit.txt", r.GetProhibitFileName())
	require.Equal(t, []string{"en/hunspell/prohibit_custom.txt"}, r.GetAdditionalProhibitFileNames())
	add := r.GetAdditionalSpellingFileNames()
	require.Contains(t, add, "en/hunspell/spelling_custom.txt")
	require.Contains(t, add, "spelling_global.txt")
	require.Contains(t, add, "/en/multiwords.txt")
	require.Equal(t, "en/hunspell/spelling_en-US.txt", r.GetLanguageVariantSpellingFileName())
}

func TestSpellingCheckRule_PathGetters_OverrideFn(t *testing.T) {
	r := NewSpellingCheckRule("S", "d", "pt-BR")
	r.GetIgnoreFileNameFn = func() string { return "pt/ignore.txt" }
	r.GetSpellingFileNameFn = func() string { return "pt/spelling.txt" }
	r.GetProhibitFileNameFn = func() string { return "pt/prohibit.txt" }
	require.Equal(t, "pt/ignore.txt", r.GetIgnoreFileName())
	require.Equal(t, "pt/spelling.txt", r.GetSpellingFileName())
	require.Equal(t, "pt/prohibit.txt", r.GetProhibitFileName())
	// PT additional (Java MorfologikPortugueseSpellerRule) replaces default list
	add := r.GetAdditionalSpellingFileNames()
	require.Contains(t, add, "spelling_global.txt")
	require.Contains(t, add, "pt/spelling.txt")
	require.Contains(t, add, "pt/multiwords.txt")
	require.NotContains(t, add, "pt/hunspell/spelling_custom.txt")
}

func TestApplyDefault_UsesPathGetters(t *testing.T) {
	// PT ignore via override path (Discover may find pt/ignore.txt)
	r := NewSpellingCheckRule("S", "d", "pt")
	r.GetIgnoreFileNameFn = func() string { return "pt/ignore.txt" }
	ApplyDefaultSpellingWordLists(r)
	// Should not panic; if pt ignore exists, words loaded
	_ = r.IgnoreWords
}
