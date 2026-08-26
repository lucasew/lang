package fr

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

func TestMorfologikFrenchSpellerRule_IDAndDict(t *testing.T) {
	r := NewMorfologikFrenchSpellerRule()
	require.Equal(t, "FR_SPELLING_RULE", r.GetID())
	require.Equal(t, "/fr/french.dict", r.GetFileName())
	// Java setIgnoreTaggedWords()
	require.True(t, r.IgnoreTaggedWords)
}

func TestMorfologikFrenchSpellerRule_IgnoreTagged(t *testing.T) {
	sp := morfologik.NewMorfologikSpeller(FrenchSpellerDict, 1)
	sp.AddWord("maison")
	r := NewMorfologikFrenchSpellerRule()
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled
	// tagged unknown token should be skipped
	sent := languagetool.AnalyzePlain("xyzzy")
	for _, tok := range sent.GetTokensWithoutWhitespace() {
		if tok == nil || tok.IsSentenceStart() {
			continue
		}
		// inject a POS tag so IsTagged
		pos := "N"
		tok.AddReading(languagetool.NewAnalyzedToken(tok.GetToken(), &pos, nil), "test")
		require.True(t, tok.IsTagged())
	}
	m, err := r.Match(sent)
	require.NoError(t, err)
	require.Empty(t, m)
}

func TestOrderFrenchSuggestions_Drops(t *testing.T) {
	got := orderFrenchSuggestions([]string{
		"foo",
		"anti virus", // prefix drop
		"x maison",   // single letter first
		"informè",    // è drop
		"burkinabè",  // è exception keep
		"le chat",    // TOKEN_AT_START → near front
	}, "zzzz")
	require.NotContains(t, got, "anti virus")
	require.NotContains(t, got, "x maison")
	require.NotContains(t, got, "informè")
	require.Contains(t, got, "burkinabè")
	require.Contains(t, got, "foo")
	require.Equal(t, "le chat", got[0])
}

func TestAdditionalTopFrenchSuggestions(t *testing.T) {
	require.Equal(t, []string{"voulais", "voulait"}, additionalTopFrenchSuggestions("voulai"))
	require.Empty(t, additionalTopFrenchSuggestions("Voulai"))
	require.Equal(t, []string{"m²"}, additionalTopFrenchSuggestions("m2"))
	require.Equal(t, []string{"km³"}, additionalTopFrenchSuggestions("KM3"))
}

func TestMatch_FrenchAdditionalTop(t *testing.T) {
	sp := morfologik.NewMorfologikSpeller(FrenchSpellerDict, 1)
	sp.AddWord("maison")
	r := NewMorfologikFrenchSpellerRule()
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled
	sent := languagetool.AnalyzePlain("voulai")
	m, err := r.Match(sent)
	require.NoError(t, err)
	require.NotEmpty(t, m)
	require.Equal(t, "voulais", m[0].GetSuggestedReplacements()[0])
}

func TestFindSuggestion_ApostropheVerb(t *testing.T) {
	r := NewMorfologikFrenchSpellerRule()
	// without TagPOS → fail-closed empty
	require.Empty(t, r.apostropheHyphenTopSuggestions("larrive"))
	// inject POS for "arrive" as verb ind/sub
	r.TagPOS = func(word string) []string {
		if word == "arrive" {
			return []string{"V ind pres 3 s"}
		}
		return nil
	}
	got := r.apostropheHyphenTopSuggestions("larrive")
	require.Equal(t, []string{"l'arrive"}, got)
}

func TestDigitSplitTopSuggestion(t *testing.T) {
	r := NewMorfologikFrenchSpellerRule()
	require.Empty(t, r.digitSplitTopSuggestion("maison2"))
	r.TagPOS = func(word string) []string {
		if word == "maison" {
			return []string{"N f s"}
		}
		return nil
	}
	require.Equal(t, "maison 2", r.digitSplitTopSuggestion("maison2"))
	// short first part only if in SPLIT_DIGITS_AT_END
	r.TagPOS = func(word string) []string {
		if word == "de" || word == "xx" {
			return []string{"P"}
		}
		return nil
	}
	require.Equal(t, "de 2", r.digitSplitTopSuggestion("de2"))
	require.Empty(t, r.digitSplitTopSuggestion("xx2"))
}

func TestSplitDigitsAtEnd(t *testing.T) {
	require.Equal(t, []string{"maison", "12"}, splitDigitsAtEnd("maison12"))
	require.Equal(t, []string{"abc"}, splitDigitsAtEnd("abc"))
}
