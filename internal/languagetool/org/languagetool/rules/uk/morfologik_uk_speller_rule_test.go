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
	require.Equal(t, UkrainianSpellerDict, NewMorfologikUkrainianSpellerRule().GetFileName())
	require.False(t, NewMorfologikUkrainianSpellerRule().IsLatinScript())
}

func TestUKIsMisspelled_TrailingHyphen(t *testing.T) {
	r := NewMorfologikUkrainianSpellerRule()
	// ends with - and not starts with - → misspelled (override, before dict)
	require.True(t, r.IsMisspelled("слово-"))
	// starts and ends with - (infix notation -ськ-) → not misspelled by this arm
	require.False(t, r.IsMisspelled("-ськ-"))
}

func TestUKIgnoreToken_NonUkrainianLetters(t *testing.T) {
	r := NewMorfologikUkrainianSpellerRule()
	sent := languagetool.AnalyzePlain("The Beatles")
	toks := sent.GetTokensWithoutWhitespace()
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
	prefs := loadDashPrefixesSpeller()
	require.NotEmpty(t, prefs)
	require.Contains(t, prefs, "блок")
	got := additionalDashPrefixSuggestions("блоксистема")
	require.Contains(t, got, "блок-система")
}

func TestDiscoverUKSpellerDict(t *testing.T) {
	p := morfologik.DiscoverLanguageDict(UkrainianSpellerDict)
	if p == "" {
		t.Skip("uk_UA.dict not discoverable")
	}
	r := NewMorfologikUkrainianSpellerRule()
	require.NotNil(t, r.Speller1, "InitSpellersFromGetters must attach multi when dict present")
	require.True(t, r.dictActive())
}
