package morfologik

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Twin of Java MorfologikSpeller(en_US.dict) isMisspelled with real FSA + .info.
func TestMorfologikSpeller_BinaryEnUS(t *testing.T) {
	path := DiscoverLanguageDict("/en/hunspell/en_US.dict")
	if path == "" {
		t.Skip("en_US.dict not in tree")
	}
	sp := NewMorfologikSpeller("/en/hunspell/en_US.dict", 1)
	require.True(t, sp.AttachBinaryDictionary(path))
	require.True(t, sp.HasDictionary())
	// en_US.info: ignore-camel-case=false, ignore-all-uppercase=false; numbers default true
	require.True(t, sp.IgnoreNumbers)
	require.False(t, sp.IgnoreCamelCase)
	require.False(t, sp.IgnoreAllUppercase)

	require.False(t, sp.IsMisspelled("software"))
	require.False(t, sp.IsMisspelled("behavior"))
	require.False(t, sp.IsMisspelled("Water")) // convertCase → water
	require.False(t, sp.IsMisspelled("WATER"))
	// ignore-numbers
	require.False(t, sp.IsMisspelled("175ºC"))
	require.False(t, sp.IsMisspelled("0º"))
	require.False(t, sp.IsMisspelled("123454"))
	// true misspellings
	require.True(t, sp.IsMisspelled("sdadsadas"))
	require.True(t, sp.IsMisspelled("bicylce"))
	require.True(t, sp.IsMisspelled("aõh"))
}

func TestMorfologikSpeller_LoadInfoFromClasspath_EN(t *testing.T) {
	if DiscoverLanguageDict("/en/hunspell/en_US.dict") == "" {
		t.Skip("en_US.dict not in tree")
	}
	// Fresh speller starts with defaults then LoadInfoFromClasspath in New
	sp := NewMorfologikSpeller("/en/hunspell/en_US.dict", 1)
	// New already tried LoadInfoFromClasspath
	require.False(t, sp.IgnoreCamelCase)
	require.False(t, sp.IgnoreAllUppercase)
	require.True(t, sp.IgnoreNumbers)
}

func TestMorfologikSpeller_LoadInfo_PolishNumbers(t *testing.T) {
	path := DiscoverLanguageDict("/pl/hunspell/pl_PL.dict")
	if path == "" {
		t.Skip("pl_PL.dict not in tree")
	}
	sp := NewMorfologikSpeller("/pl/hunspell/pl_PL.dict", 1)
	require.True(t, sp.LoadInfoBesideDict(path))
	// pl_PL.info: ignore-numbers=false
	require.False(t, sp.IgnoreNumbers)
	require.False(t, sp.IgnoreCamelCase)
	require.False(t, sp.IgnoreAllUppercase)
}

func TestMorfologikSpellerRule_BinaryMatch(t *testing.T) {
	path := DiscoverLanguageDict("/en/hunspell/en_US.dict")
	if path == "" {
		t.Skip("en_US.dict not in tree")
	}
	sp := NewMorfologikSpeller("/en/hunspell/en_US.dict", 1)
	require.True(t, sp.AttachBinaryDictionary(path))
	r := NewMorfologikSpellerRule("MORFOLOGIK_RULE_EN_US", "en", "/en/hunspell/en_US.dict", sp)
	if r.SpellingCheckRule != nil {
		r.IgnoreWordsWithLength = 1
	}
	// known sentence — Water freezes accepted via convertCase + freezes in dict
	ms, err := r.Match(languagetool.AnalyzePlain("Water freezes."))
	require.NoError(t, err)
	require.Empty(t, ms)
	// misspelling
	ms, err = r.Match(languagetool.AnalyzePlain("bicylce"))
	require.NoError(t, err)
	require.NotEmpty(t, ms)
}
