package de

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// testGender ports getNounGender for unit tests only (not production invent).
func testGender(w string) string {
	switch w {
	case "Mann", "Plan", "Tisch", "Baum", "Hund", "Tag", "Name":
		return "MAS"
	case "Frau", "Idee", "Roadmap", "Katze", "Zeit", "Stadt", "Blume":
		return "FEM"
	case "Kind", "Verfahren", "Haus", "Buch", "Auto", "Mädchen", "Tier":
		return "NEU"
	default:
		return ""
	}
}

// testSynthesize ports a minimal GermanSynthesizer subset for ART/PRO used in tests.
// Keys are built from postagRE containing CASE and GENDER after MAS|FEM|NEU replace.
func testSynthesize(lemma, postagRE string) []string {
	// postagRE examples: ART:DEF:NOM:SIN:MAS, ART:IND:AKK:SIN:MAS, PRO:POS:NOM:SIN:FEM
	gender := ""
	for _, g := range []string{"MAS", "FEM", "NEU"} {
		if strings.Contains(postagRE, g) {
			gender = g
			break
		}
	}
	cas := ""
	for _, c := range []string{"NOM", "AKK", "GEN", "DAT"} {
		if strings.Contains(postagRE, c) {
			cas = c
			break
		}
	}
	key := strings.ToLower(lemma) + "|" + cas + "|" + gender
	if forms, ok := testDetSynth[key]; ok {
		return append([]string(nil), forms...)
	}
	return nil
}

var testDetSynth = map[string][]string{
	"der|NOM|MAS": {"der"}, "der|AKK|MAS": {"den"}, "der|GEN|MAS": {"des"}, "der|DAT|MAS": {"dem"},
	"der|NOM|FEM": {"die"}, "der|AKK|FEM": {"die"}, "der|GEN|FEM": {"der"}, "der|DAT|FEM": {"der"},
	"der|NOM|NEU": {"das"}, "der|AKK|NEU": {"das"}, "der|GEN|NEU": {"des"}, "der|DAT|NEU": {"dem"},
	"ein|NOM|MAS": {"ein"}, "ein|AKK|MAS": {"einen"}, "ein|GEN|MAS": {"eines"}, "ein|DAT|MAS": {"einem"},
	"ein|NOM|FEM": {"eine"}, "ein|AKK|FEM": {"eine"}, "ein|GEN|FEM": {"einer"}, "ein|DAT|FEM": {"einer"},
	"ein|NOM|NEU": {"ein"}, "ein|AKK|NEU": {"ein"}, "ein|GEN|NEU": {"eines"}, "ein|DAT|NEU": {"einem"},
	"mein|NOM|MAS": {"mein"}, "mein|AKK|MAS": {"meinen"}, "mein|GEN|MAS": {"meines"}, "mein|DAT|MAS": {"meinem"},
	"mein|NOM|FEM": {"meine"}, "mein|AKK|FEM": {"meine"}, "mein|GEN|FEM": {"meiner"}, "mein|DAT|FEM": {"meiner"},
	"mein|NOM|NEU": {"mein"}, "mein|AKK|NEU": {"mein"}, "mein|GEN|NEU": {"meines"}, "mein|DAT|NEU": {"meinem"},
	// Possessives share Java synthesizer paradigms; first-char filter keeps surface lemma family.
	"dein|NOM|MAS": {"dein"}, "dein|AKK|MAS": {"deinen"}, "dein|GEN|MAS": {"deines"}, "dein|DAT|MAS": {"deinem"},
	"dein|NOM|FEM": {"deine"}, "dein|AKK|FEM": {"deine"}, "dein|GEN|FEM": {"deiner"}, "dein|DAT|FEM": {"deiner"},
	"dein|NOM|NEU": {"dein"}, "dein|AKK|NEU": {"dein"}, "dein|GEN|NEU": {"deines"}, "dein|DAT|NEU": {"deinem"},
	"sein|NOM|MAS": {"sein"}, "sein|AKK|MAS": {"seinen"}, "sein|GEN|MAS": {"seines"}, "sein|DAT|MAS": {"seinem"},
	"sein|NOM|FEM": {"seine"}, "sein|AKK|FEM": {"seine"}, "sein|GEN|FEM": {"seiner"}, "sein|DAT|FEM": {"seiner"},
	"sein|NOM|NEU": {"sein"}, "sein|AKK|NEU": {"sein"}, "sein|GEN|NEU": {"seines"}, "sein|DAT|NEU": {"seinem"},
	"ihr|NOM|MAS": {"ihr"}, "ihr|AKK|MAS": {"ihren"}, "ihr|GEN|MAS": {"ihres"}, "ihr|DAT|MAS": {"ihrem"},
	"ihr|NOM|FEM": {"ihre"}, "ihr|AKK|FEM": {"ihre"}, "ihr|GEN|FEM": {"ihrer"}, "ihr|DAT|FEM": {"ihrer"},
	"ihr|NOM|NEU": {"ihr"}, "ihr|AKK|NEU": {"ihr"}, "ihr|GEN|NEU": {"ihres"}, "ihr|DAT|NEU": {"ihrem"},
	"unser|NOM|MAS": {"unser"}, "unser|AKK|MAS": {"unseren"}, "unser|GEN|MAS": {"unseres"}, "unser|DAT|MAS": {"unserem"},
	"unser|NOM|FEM": {"unsere"}, "unser|AKK|FEM": {"unsere"}, "unser|GEN|FEM": {"unserer"}, "unser|DAT|FEM": {"unserer"},
	"unser|NOM|NEU": {"unser"}, "unser|AKK|NEU": {"unser"}, "unser|GEN|NEU": {"unseres"}, "unser|DAT|NEU": {"unserem"},
	"euer|NOM|MAS": {"euer"}, "euer|AKK|MAS": {"euren"}, "euer|GEN|MAS": {"eures"}, "euer|DAT|MAS": {"eurem"},
	"euer|NOM|FEM": {"eure"}, "euer|AKK|FEM": {"eure"}, "euer|GEN|FEM": {"eurer"}, "euer|DAT|FEM": {"eurer"},
	"euer|NOM|NEU": {"euer"}, "euer|AKK|NEU": {"euer"}, "euer|GEN|NEU": {"eures"}, "euer|DAT|NEU": {"eurem"},
}

func withTestHooks(f *AdaptSuggestionFilter) *AdaptSuggestionFilter {
	f.GenderOf = testGender
	f.Synthesize = testSynthesize
	return f
}

func TestAdaptSuggestionFilter_FailClosedWithoutHooks(t *testing.T) {
	f := NewAdaptSuggestionFilter()
	// soft invent removed: no GenderOf/Synthesize → empty
	require.Empty(t, f.AdaptedDet(DetReading{Token: "die", POS: "ART:DEF:NOM:SIN:FEM", Lemma: "der"}, "Mann"))
	require.Empty(t, f.SuggestWithDet("die", "ART:DEF:NOM:SIN:FEM", "der", []string{"Plan"}))
}

func TestAdaptSuggestionFilter_AdaptedDet(t *testing.T) {
	f := withTestHooks(NewAdaptSuggestionFilter())
	require.Equal(t, []string{"der"}, f.AdaptedDet(DetReading{Token: "die", POS: "ART:DEF:NOM:SIN:FEM", Lemma: "der"}, "Mann"))
	require.Equal(t, []string{"die"}, f.AdaptedDet(DetReading{Token: "der", POS: "ART:DEF:NOM:SIN:MAS", Lemma: "der"}, "Frau"))
	require.Equal(t, []string{"das"}, f.AdaptedDet(DetReading{Token: "der", POS: "ART:DEF:NOM:SIN:NEU", Lemma: "der"}, "Kind"))

	require.Equal(t, []string{"ein"}, f.AdaptedDet(DetReading{Token: "eine", POS: "ART:IND:NOM:SIN:FEM", Lemma: "ein"}, "Plan"))
	require.Equal(t, []string{"einen"}, f.AdaptedDet(DetReading{Token: "eine", POS: "ART:IND:AKK:SIN:FEM", Lemma: "ein"}, "Plan"))
	require.Equal(t, []string{"eines"}, f.AdaptedDet(DetReading{Token: "einer", POS: "ART:IND:GEN:SIN:FEM", Lemma: "ein"}, "Plan"))
	require.Equal(t, []string{"einem"}, f.AdaptedDet(DetReading{Token: "einer", POS: "ART:IND:DAT:SIN:FEM", Lemma: "ein"}, "Plan"))
}

func TestAdaptSuggestionFilter_Possessives(t *testing.T) {
	f := withTestHooks(NewAdaptSuggestionFilter())
	require.Equal(t, []string{"mein"}, f.AdaptedDet(DetReading{Token: "meine", POS: "PRO:POS:NOM:SIN:FEM", Lemma: "mein"}, "Plan"))
	require.Equal(t, []string{"meinen"}, f.AdaptedDet(DetReading{Token: "meine", POS: "PRO:POS:AKK:SIN:FEM", Lemma: "mein"}, "Plan"))
	require.Equal(t, []string{"unser"}, f.AdaptedDet(DetReading{Token: "unsere", POS: "PRO:POS:NOM:SIN:FEM", Lemma: "unser"}, "Plan"))
	require.Equal(t, []string{"unseren"}, f.AdaptedDet(DetReading{Token: "unsere", POS: "PRO:POS:AKK:SIN:FEM", Lemma: "unser"}, "Plan"))
	require.Equal(t, []string{"das"}, f.AdaptedDet(DetReading{Token: "die", POS: "ART:DEF:NOM:SIN:FEM", Lemma: "der"}, "Verfahren"))
}

func TestAdaptSuggestionFilter_SuggestWithDet(t *testing.T) {
	f := withTestHooks(NewAdaptSuggestionFilter())
	got := f.SuggestWithDet("die", "ART:DEF:NOM:SIN:FEM", "der", []string{"Plan"})
	require.Equal(t, []string{"der Plan"}, got)

	got = f.SuggestWithDet("meine", "PRO:POS:NOM:SIN:FEM", "mein", []string{"Idee"})
	require.Equal(t, []string{"meine Idee"}, got)
}

func TestAdaptSuggestionFilter_UppercaseDet(t *testing.T) {
	f := withTestHooks(NewAdaptSuggestionFilter())
	got := f.AdaptedDet(DetReading{Token: "Die", POS: "ART:DEF:NOM:SIN:FEM", Lemma: "der"}, "Plan")
	require.Equal(t, []string{"Der"}, got)
}
