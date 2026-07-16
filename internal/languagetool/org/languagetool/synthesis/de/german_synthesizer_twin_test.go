package de

import (
	"strings"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
	"github.com/stretchr/testify/require"
)

func TestGermanSynthesizer_Synthesize(t *testing.T) {
	manual, err := synthesis.NewManualSynthesizer(strings.NewReader(strings.Join([]string{
		"Häuser\tHaus\tSUB:AKK:PLU:NEU",
		"Hauses\tHaus\tSUB:GEN:SIN:NEU",
		"Zug\tZug\tSUB:DAT:SIN:MAS",
		"Tisch\tTisch\tSUB:DAT:SIN:MAS",
	}, "\n") + "\n"))
	require.NoError(t, err)
	s := NewGermanSynthesizer(manual)

	lemma := "Haus"
	tok := languagetool.NewAnalyzedToken("Haus", nil, &lemma)
	forms, err := s.Synthesize(tok, "SUB:AKK:PLU:NEU")
	require.NoError(t, err)
	require.Equal(t, []string{"Häuser"}, forms)

	forms, err = s.Synthesize(tok, "SUB:NOM:PLU:MAS")
	require.NoError(t, err)
	require.Empty(t, forms)

	zlemma := "Zug"
	ztok := languagetool.NewAnalyzedToken("Zug", nil, &zlemma)
	forms, err = s.Synthesize(ztok, "SUB:DAT:SIN:MAS")
	require.NoError(t, err)
	require.Equal(t, []string{"Zug"}, forms)

	forms, err = s.Synthesize(languagetool.NewAnalyzedToken("fake", nil, nil), "FAKE")
	require.NoError(t, err)
	require.Empty(t, forms)
}

func TestGermanSynthesizer_SynthesizeCompounds(t *testing.T) {
	manual, err := synthesis.NewManualSynthesizer(strings.NewReader(
		"Regelsysteme\tRegelsystem\tSUB:NOM:PLU:NEU\n" +
			"Regelsystemen\tRegelsystem\tSUB:DAT:PLU:NEU\n"))
	require.NoError(t, err)
	s := NewGermanSynthesizer(manual)
	lemma := "Regelsystem"
	tok := languagetool.NewAnalyzedToken("Regelsystem", nil, &lemma)
	forms, err := s.Synthesize(tok, "SUB:NOM:PLU:NEU")
	require.NoError(t, err)
	require.Equal(t, []string{"Regelsysteme"}, forms)
	forms, err = s.SynthesizeRE(tok, ".*:PLU:.*", true)
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"Regelsysteme", "Regelsystemen"}, forms)
}

func TestGermanSynthesizer_MorfologikBug(t *testing.T) {
	manual, err := synthesis.NewManualSynthesizer(strings.NewReader(
		"anfragen\tanfragen\tVER:1:PLU:KJ1:SFT:NEB\n"))
	require.NoError(t, err)
	s := NewGermanSynthesizer(manual)
	lemma := "anfragen"
	tok := languagetool.NewAnalyzedToken("anfragen", nil, &lemma)
	forms, err := s.Synthesize(tok, "VER:1:PLU:KJ1:SFT:NEB")
	require.NoError(t, err)
	require.Equal(t, []string{"anfragen"}, forms)
}
