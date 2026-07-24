package uk

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

func TestMorfologikUkrainianSpellerRule_ID(t *testing.T) {
	require.Equal(t, MorfologikUkrainianSpellerRuleID, NewMorfologikUkrainianSpellerRule().GetID())
	require.Equal(t, "/uk/hunspell/uk_UA.dict", UkrainianSpellerDict)
}

func TestUKIsMisspelled_TrailingHyphen(t *testing.T) {
	r := NewMorfologikUkrainianSpellerRule()
	// ends with - and not starts with - → misspelled
	require.True(t, r.IsMisspelled("слово-"))
	// starts and ends with - (infix notation -ськ-) → not misspelled by this arm
	require.False(t, r.IsMisspelled("-ськ-"))
}

func TestUKIgnoreToken_NonUkrainianLetters(t *testing.T) {
	r := NewMorfologikUkrainianSpellerRule()
	sent := languagetool.AnalyzePlain("The Beatles")
	toks := sent.GetTokensWithoutWhitespace()
	// find "The"
	for i, tok := range toks {
		if tok == nil || tok.IsSentenceStart() {
			continue
		}
		if tok.GetToken() == "The" || tok.GetToken() == "Beatles" {
			require.True(t, r.ukIgnoreToken(toks, i), tok.GetToken())
		}
	}
}

func TestUKIgnoreToken_HasGoodTag(t *testing.T) {
	r := NewMorfologikUkrainianSpellerRule()
	sent := languagetool.AnalyzePlain("слово")
	toks := sent.GetTokensWithoutWhitespace()
	for i, tok := range toks {
		if tok == nil || tok.IsSentenceStart() {
			continue
		}
		// without tags → hasGoodTag false → not ignored via tag arm
		require.False(t, hasGoodTagUK(tok))
		pos := "noun:inanim:n:v_naz"
		tok.AddReading(languagetool.NewAnalyzedToken(tok.GetToken(), &pos, nil), "test")
		require.True(t, hasGoodTagUK(tok))
		require.True(t, r.ukIgnoreToken(toks, i))
	}
}

func TestFilterUKSuggestions(t *testing.T) {
	got := filterUKSuggestions([]string{"кіно прокат", "нормальне", "вело- прогулянка", "супер слово"})
	require.Equal(t, []string{"нормальне"}, got)
}

func TestAdditionalDashPrefixSuggestions(t *testing.T) {
	// need a prefix from dash_prefixes that is ≥3 cyrillic and not :alt/:bad/:slang
	prefs := loadDashPrefixesSpeller()
	require.NotEmpty(t, prefs)
	// pick any key
	var key string
	for k := range prefs {
		key = k
		break
	}
	// word = key + "абвгд" without hyphen
	word := key + "абвгд"
	got := additionalDashPrefixSuggestions(word)
	require.NotEmpty(t, got)
	require.Contains(t, got, key+"-"+"абвгд")
}

func TestMatch_NonUkrainianIgnored(t *testing.T) {
	sp := morfologik.NewMorfologikSpeller(UkrainianSpellerDict, 1)
	sp.AddWord("слово")
	r := NewMorfologikUkrainianSpellerRule()
	r.Speller = sp
	// re-wrap IsMisspelled for hyphen + map
	inner := sp.IsMisspelled
	r.IsMisspelled = func(w string) bool { return r.ukIsMisspelled(w, inner) }
	sent := languagetool.AnalyzePlain("The Beatles")
	m, err := r.Match(sent)
	require.NoError(t, err)
	require.Empty(t, m)
}

func TestMatch_MisspellUkrainian(t *testing.T) {
	sp := morfologik.NewMorfologikSpeller(UkrainianSpellerDict, 1)
	sp.AddWord("слово")
	r := NewMorfologikUkrainianSpellerRule()
	r.Speller = sp
	inner := sp.IsMisspelled
	r.IsMisspelled = func(w string) bool { return r.ukIsMisspelled(w, inner) }
	sent := languagetool.AnalyzePlain("слвво")
	m, err := r.Match(sent)
	require.NoError(t, err)
	require.Len(t, m, 1)
}

func TestMatch_HasGoodTagSkipped(t *testing.T) {
	sp := morfologik.NewMorfologikSpeller(UkrainianSpellerDict, 1)
	// empty of "слвво" — would flag unless tagged
	r := NewMorfologikUkrainianSpellerRule()
	r.Speller = sp
	inner := sp.IsMisspelled
	r.IsMisspelled = func(w string) bool { return r.ukIsMisspelled(w, inner) }
	sent := languagetool.AnalyzePlain("слвво")
	for _, tok := range sent.GetTokensWithoutWhitespace() {
		if tok == nil || tok.IsSentenceStart() {
			continue
		}
		pos := "noun:inanim"
		tok.AddReading(languagetool.NewAnalyzedToken(tok.GetToken(), &pos, nil), "test")
	}
	m, err := r.Match(sent)
	require.NoError(t, err)
	require.Empty(t, m)
}
