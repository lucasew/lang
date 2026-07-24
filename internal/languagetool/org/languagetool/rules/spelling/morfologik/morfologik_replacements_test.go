package morfologik

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestApplyConversionPairs(t *testing.T) {
	// ordered sequential replaces
	pairs := [][2]string{{"æ", "ae"}, {"œ", "oe"}}
	require.Equal(t, "aether", ApplyConversionPairs("æther", pairs))
	require.Equal(t, "nochange", ApplyConversionPairs("nochange", pairs))
	require.Equal(t, "x", ApplyConversionPairs("x", nil))
}

func TestParseConversionPairs(t *testing.T) {
	p := ParseConversionPairs("a b, c d, a z")
	// first a wins
	require.Equal(t, [][2]string{{"a", "b"}, {"c", "d"}}, p)
}

func TestParseAndPartitionReplacementPairs(t *testing.T) {
	// from en_US-style: short targets → short list; long → theRest
	raw := "f ph, lite light, shun tion, a ei"
	pairs := ParseReplacementPairs(raw)
	require.GreaterOrEqual(t, len(pairs), 4)
	rest, short := partitionReplacementPairs(pairs)
	// lite→light (5), shun→tion (4) kept in theRest; f→ph and a→ei in short
	require.Contains(t, rest.Get("lite"), "light")
	require.Contains(t, rest.Get("shun"), "tion")
	require.Nil(t, rest.Get("f"))
	require.Nil(t, rest.Get("a"))
	// LinkedHashMap order: lite before shun (f/a went to short)
	require.Equal(t, []string{"lite", "shun"}, rest.Keys)
	require.NotEmpty(t, short)
	foundF := false
	for _, p := range short {
		if p.From == "f" && p.To == "ph" {
			foundF = true
		}
	}
	require.True(t, foundF, "short=%v", short)
}

func TestGetAllReplacements_LiteLight(t *testing.T) {
	rest := &LinkedHashStringListMap{}
	rest.Add("lite", "light")
	got := GetAllReplacements("lite", rest, 0, 0)
	require.Contains(t, got, "lite")  // branch without replacement
	require.Contains(t, got, "light") // branch with
}

func TestParseReplacementPairs_UnderscoreSpace(t *testing.T) {
	// Java hunspell REP: '_' → space
	p := ParseReplacementPairs("a_b c_d")
	require.Equal(t, []ReplacementPair{{From: "a b", To: "c d"}}, p)
}

func TestPartitionReplacementPairs_LinkedHashOrder(t *testing.T) {
	pairs := ParseReplacementPairs("lite light, shun tion, phoby phobia")
	rest, short := partitionReplacementPairs(pairs)
	require.Empty(t, short)
	require.Equal(t, []string{"lite", "shun", "phoby"}, rest.Keys)
}

func TestBinaryReplacementPairs_Phoby(t *testing.T) {
	path := DiscoverLanguageDict("/en/hunspell/en_US.dict")
	if path == "" {
		t.Skip("en_US.dict not in tree")
	}
	sp := NewMorfologikSpeller("/en/hunspell/en_US.dict", 1)
	require.True(t, sp.AttachBinaryDictionary(path))
	require.NotEmpty(t, sp.ReplacementTheRest, "en_US.info replacement-pairs should load")
	// "lite" is itself in the dict; use misspelled "phoby" → "phobia" (pair phoby phobia).
	require.True(t, sp.IsMisspelled("phoby"))
	sugs := sp.GetWeightedSuggestions("phoby")
	require.NotEmpty(t, sugs)
	require.Equal(t, "phobia", sugs[0].Word, "distance-0 replacement should rank first; sugs=%v", sugs)
	// distance 0 weight = 26 - freq - 1 < typical edit-1 weight 51-freq
	require.Less(t, sugs[0].Weight, 51)

	// shun→tion multi-char: attenshun → attention
	sugs2 := sp.FindReplacements("attenshun")
	require.Contains(t, sugs2, "attention")
}

func TestInputConversion_IsMisspelled(t *testing.T) {
	sp := NewMorfologikSpeller("/xx.dict", 1)
	sp.AddWord("ae")
	sp.InputConversionPairs = [][2]string{{"æ", "ae"}}
	// æ converts to ae which is in dict
	require.False(t, sp.IsMisspelled("æ"))
	require.True(t, sp.IsMisspelled("ø"))
}
func TestBinaryShortReplacementPairs_EN(t *testing.T) {
	path := DiscoverLanguageDict("/en/hunspell/en_US.dict")
	if path == "" {
		t.Skip("en_US.dict not in tree")
	}
	sp := NewMorfologikSpeller("/en/hunspell/en_US.dict", 1)
	require.True(t, sp.AttachBinaryDictionary(path))
	require.NotEmpty(t, sp.ReplacementShort, "short f/ph etc. pairs")
	// f→ph / kw→qu only via findRepl anyToOne/anyToTwo (Java Speller 2.2.0), not surface invent.
	require.True(t, sp.IsMisspelled("fone"))
	sugs := sp.FindReplacements("fone")
	require.Contains(t, sugs, "phone", "sugs=%v", sugs)
	// weighted: pure replacement should be distance 0 (weight < typical edit-1 ≈ 51-freq)
	ws := sp.GetWeightedSuggestions("fone")
	require.NotEmpty(t, ws)
	foundPhone := false
	for _, w := range ws {
		if w.Word == "phone" {
			foundPhone = true
			require.Less(t, w.Weight, 51, "phone should be dist-0 class weight; got %d", w.Weight)
		}
	}
	require.True(t, foundPhone, "ws=%v", ws)
	require.True(t, sp.IsMisspelled("kwality"))
	sugs2 := sp.FindReplacements("kwality")
	require.Contains(t, sugs2, "quality", "sugs=%v", sugs2)
}
