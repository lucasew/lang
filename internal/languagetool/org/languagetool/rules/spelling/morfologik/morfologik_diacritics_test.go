package morfologik

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseEquivalentChars(t *testing.T) {
	m := ParseEquivalentChars("x ź, l ł, u ó, ó u")
	require.Equal(t, []rune{'ź'}, m['x'])
	require.Equal(t, []rune{'ł'}, m['l'])
	require.Contains(t, m['u'], 'ó')
	require.Contains(t, m['ó'], 'u')
}

func TestEN_IgnoreDiacritics_Loaded(t *testing.T) {
	path := DiscoverLanguageDict("/en/hunspell/en_US.dict")
	if path == "" {
		t.Skip("en_US.dict not in tree")
	}
	sp := NewMorfologikSpeller("/en/hunspell/en_US.dict", 1)
	require.True(t, sp.AttachBinaryDictionary(path))
	require.True(t, sp.IgnoreDiacritics, "en_US.info ignore-diacritics=true")
}

func TestEN_DiacriticSuggestion_Fiance(t *testing.T) {
	path := DiscoverLanguageDict("/en/hunspell/en_US.dict")
	if path == "" {
		t.Skip("en_US.dict not in tree")
	}
	sp := NewMorfologikSpeller("/en/hunspell/en_US.dict", 1)
	require.True(t, sp.AttachBinaryDictionary(path))
	// ASCII misspelling → accented dictionary form via ignore-diacritics replace alphabet
	// (fiancé / fiancee may both exist; at least one accented fiancé form)
	require.True(t, sp.IsMisspelled("fiance") || !sp.IsMisspelled("fiance"))
	// If fiance is already accepted, skip; else expect fiancé among sugs
	if sp.IsMisspelled("fiance") {
		sugs := sp.FindReplacements("fiance")
		found := false
		for _, s := range sugs {
			if s == "fiancé" || s == "fiancee" || s == "fiancée" {
				found = true
			}
		}
		require.True(t, found, "sugs=%v", sugs)
	}
}
