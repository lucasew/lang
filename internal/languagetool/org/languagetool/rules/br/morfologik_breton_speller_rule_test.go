package br

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/stretchr/testify/require"
)

func TestMorfologikBretonSpellerRule(t *testing.T) {
	r := NewMorfologikBretonSpellerRule()
	require.Equal(t, MorfologikBretonSpellerRuleID, r.GetID())
	require.Equal(t, MorfologikBretonSpellerRuleDict, r.GetFileName())
	// Java setIgnoreTaggedWords()
	require.True(t, r.IgnoreTaggedWords)
}

func TestBretonIgnoreTagged(t *testing.T) {
	sp := morfologik.NewMorfologikSpeller(MorfologikBretonSpellerRuleDict, 1)
	sp.AddWord("test")
	r := NewMorfologikBretonSpellerRule()
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled
	sent := languagetool.AnalyzePlain("xyzzy")
	for _, tok := range sent.GetTokensWithoutWhitespace() {
		if tok == nil || tok.IsSentenceStart() {
			continue
		}
		pos := "N"
		tok.AddReading(languagetool.NewAnalyzedToken(tok.GetToken(), &pos, nil), "test")
	}
	m, err := r.Match(sent)
	require.NoError(t, err)
	require.Empty(t, m)
}

func TestBretonHyphenTokenizing(t *testing.T) {
	sp := morfologik.NewMorfologikSpeller(MorfologikBretonSpellerRuleDict, 1)
	sp.AddWord("a")
	sp.AddWord("b")
	r := NewMorfologikBretonSpellerRule()
	r.Speller = sp
	r.IsMisspelled = sp.IsMisspelled
	// whole "a-b" not in dict → parent flags; parts a,b accepted → drop
	sent := languagetool.AnalyzePlain("a-b")
	m, err := r.Match(sent)
	require.NoError(t, err)
	require.Empty(t, m)
	// part misspelled → keep
	sent = languagetool.AnalyzePlain("a-zz")
	m, err = r.Match(sent)
	require.NoError(t, err)
	require.NotEmpty(t, m)
}
