package de

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAdaptSuggestionFilter_AdaptedDet(t *testing.T) {
	f := NewAdaptSuggestionFilter()
	// Port of AdaptSuggestionFilterTest.testAdaptedDet
	require.Equal(t, []string{"der"}, f.AdaptedDet(DetReading{Token: "die", POS: "ART:DEF:NOM:SIN:FEM", Lemma: "der"}, "Mann"))
	require.Equal(t, []string{"die"}, f.AdaptedDet(DetReading{Token: "der", POS: "ART:DEF:NOM:SIN:MAS", Lemma: "der"}, "Frau"))
	require.Equal(t, []string{"das"}, f.AdaptedDet(DetReading{Token: "der", POS: "ART:DEF:NOM:SIN:NEU", Lemma: "der"}, "Kind"))

	require.Equal(t, []string{"ein"}, f.AdaptedDet(DetReading{Token: "eine", POS: "ART:IND:NOM:SIN:FEM", Lemma: "ein"}, "Plan"))
	require.Equal(t, []string{"einen"}, f.AdaptedDet(DetReading{Token: "eine", POS: "ART:IND:AKK:SIN:FEM", Lemma: "ein"}, "Plan"))
	require.Equal(t, []string{"eines"}, f.AdaptedDet(DetReading{Token: "einer", POS: "ART:IND:GEN:SIN:FEM", Lemma: "ein"}, "Plan"))
	require.Equal(t, []string{"einem"}, f.AdaptedDet(DetReading{Token: "einer", POS: "ART:IND:DAT:SIN:FEM", Lemma: "ein"}, "Plan"))
}

func TestAdaptSuggestionFilter_Possessives(t *testing.T) {
	f := NewAdaptSuggestionFilter()
	require.Equal(t, []string{"mein"}, f.AdaptedDet(DetReading{Token: "meine", POS: "PRO:POS:NOM:SIN:FEM", Lemma: "mein"}, "Plan"))
	require.Equal(t, []string{"meinen"}, f.AdaptedDet(DetReading{Token: "meine", POS: "PRO:POS:AKK:SIN:FEM", Lemma: "mein"}, "Plan"))
	require.Equal(t, []string{"unser"}, f.AdaptedDet(DetReading{Token: "unsere", POS: "PRO:POS:NOM:SIN:FEM", Lemma: "unser"}, "Plan"))
	require.Equal(t, []string{"unseren"}, f.AdaptedDet(DetReading{Token: "unsere", POS: "PRO:POS:AKK:SIN:FEM", Lemma: "unser"}, "Plan"))
	require.Equal(t, []string{"das"}, f.AdaptedDet(DetReading{Token: "die", POS: "ART:DEF:NOM:SIN:FEM", Lemma: "der"}, "Verfahren"))
}

func TestAdaptSuggestionFilter_SuggestWithDet(t *testing.T) {
	f := NewAdaptSuggestionFilter()
	// "die Roadmap" → "Plan" (MAS): der/den depending on case; NOM reading of "die" FEM → der for MAS NOM
	got := f.SuggestWithDet("die", "ART:DEF:NOM:SIN:FEM", "der", []string{"Plan"})
	require.Equal(t, []string{"der Plan"}, got)

	got = f.SuggestWithDet("meine", "PRO:POS:NOM:SIN:FEM", "mein", []string{"Idee"})
	require.Equal(t, []string{"meine Idee"}, got)
}

func TestAdaptSuggestionFilter_UppercaseDet(t *testing.T) {
	f := NewAdaptSuggestionFilter()
	got := f.AdaptedDet(DetReading{Token: "Die", POS: "ART:DEF:NOM:SIN:FEM", Lemma: "der"}, "Plan")
	require.Equal(t, []string{"Der"}, got)
}
