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

// REMOVE + case gate (Java GermanSynthesizer.lookup / synthesize).
func TestGermanSynthesizer_RemoveAndCase(t *testing.T) {
	manual, err := synthesis.NewManualSynthesizer(strings.NewReader(
		"schloß\tschließen\tVER:3:SIN:PRT:NON:NEB\n" +
			"schloss\tschließen\tVER:3:SIN:PRT:NON:NEB\n" +
			"Häuser\thaus\tSUB:NOM:PLU:NEU\n" + // uppercase form for lowercase lemma → drop
			"häuser\thaus\tSUB:NOM:PLU:NEU\n" +
			"Ihr\tmein\tPRO:PER:NOM:PLU:ALG\n", // mein allows cross-case
	))
	require.NoError(t, err)
	s := NewGermanSynthesizer(manual)

	// old spelling in REMOVE set is dropped
	lem := "schließen"
	tok := languagetool.NewAnalyzedToken("schließen", nil, &lem)
	forms, err := s.Synthesize(tok, "VER:3:SIN:PRT:NON:NEB")
	require.NoError(t, err)
	require.Equal(t, []string{"schloss"}, forms)

	// lowercase lemma must not yield uppercase form
	hl := "haus"
	htok := languagetool.NewAnalyzedToken("haus", nil, &hl)
	forms, err = s.Synthesize(htok, "SUB:NOM:PLU:NEU")
	require.NoError(t, err)
	require.Equal(t, []string{"häuser"}, forms)

	// mein allows Ihr (cross-case exception)
	ml := "mein"
	mtok := languagetool.NewAnalyzedToken("mein", nil, &ml)
	forms, err = s.Synthesize(mtok, "PRO:PER:NOM:PLU:ALG")
	require.NoError(t, err)
	require.Equal(t, []string{"Ihr"}, forms)
}

// getCompoundForms: whole compound not in synth dict → split last part, rejoin (Java).
func TestGermanSynthesizer_GetCompoundForms(t *testing.T) {
	// Only last-part lemma Boot is in the synth manual; Hausboot is not.
	manual, err := synthesis.NewManualSynthesizer(strings.NewReader(
		"Boote\tBoot\tSUB:NOM:PLU:NEU\n" +
			"Booten\tBoot\tSUB:DAT:PLU:NEU\n",
	))
	require.NoError(t, err)
	s := NewGermanSynthesizer(manual)

	lemma := "Hausboot"
	tok := languagetool.NewAnalyzedToken("Hausboot", nil, &lemma)

	// Without compound tokenizer: fail-closed (no invent split)
	forms, err := s.Synthesize(tok, "SUB:NOM:PLU:NEU")
	require.NoError(t, err)
	require.Empty(t, forms)

	// Strict split like GermanCompoundTokenizer when lexicon knows parts
	s.StrictCompoundTokenize = func(w string) []string {
		if w == "Hausboot" {
			return []string{"Haus", "Boot"}
		}
		if w == "Haus-Boot" {
			return []string{"Haus-Boot"} // single → hyphen split in getCompoundForms
		}
		return []string{w}
	}
	forms, err = s.Synthesize(tok, "SUB:NOM:PLU:NEU")
	require.NoError(t, err)
	require.Equal(t, []string{"Hausboote"}, forms)

	forms, err = s.SynthesizeRE(tok, ".*:PLU:.*", true)
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"Hausboote", "Hausbooten"}, forms)

	// Hyphen compounds: last segment keeps capital when original last part was capital
	hyLemma := "Haus-Boot"
	hyTok := languagetool.NewAnalyzedToken("Haus-Boot", nil, &hyLemma)
	forms, err = s.Synthesize(hyTok, "SUB:NOM:PLU:NEU")
	require.NoError(t, err)
	require.Equal(t, []string{"Haus-Boote"}, forms)
}
