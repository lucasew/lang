package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/AdaptSuggestionFilterTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of AdaptSuggestionFilterTest.testAcceptRuleMatchWithDet (surface det path).
func TestAdaptSuggestionFilter_AcceptRuleMatchWithDet(t *testing.T) {
	f := withTestHooks(NewAdaptSuggestionFilter())
	require.Equal(t, []string{"der Plan"}, f.SuggestWithDet("die", "ART:DEF:NOM:SIN:FEM", "der", []string{"Plan"}))
	require.Equal(t, []string{"ein Plan"}, f.SuggestWithDet("eine", "ART:IND:NOM:SIN:FEM", "ein", []string{"Plan"}))
	require.Equal(t, []string{"mein Plan"}, f.SuggestWithDet("meine", "PRO:POS:NOM:SIN:FEM", "mein", []string{"Plan"}))
	require.Equal(t, []string{"die Idee"}, f.SuggestWithDet("die", "ART:DEF:NOM:SIN:FEM", "der", []string{"Idee"}))
	require.Equal(t, []string{"meine Idee"}, f.SuggestWithDet("meine", "PRO:POS:NOM:SIN:FEM", "mein", []string{"Idee"}))
	require.Equal(t, []string{"das Verfahren"}, f.SuggestWithDet("die", "ART:DEF:NOM:SIN:FEM", "der", []string{"Verfahren"}))
	require.Equal(t, []string{"mein Verfahren"}, f.SuggestWithDet("meine", "PRO:POS:NOM:SIN:FEM", "mein", []string{"Verfahren"}))
}

// Port of AdaptSuggestionFilterTest.testAdaptedDet
func TestAdaptSuggestionFilter_AdaptedDet_Twin(t *testing.T) {
	f := withTestHooks(NewAdaptSuggestionFilter())
	require.Equal(t, []string{"der"}, f.AdaptedDet(DetReading{Token: "die", POS: "ART:DEF:NOM:SIN:FEM", Lemma: "der"}, "Mann"))
	require.Equal(t, []string{"die"}, f.AdaptedDet(DetReading{Token: "der", POS: "ART:DEF:NOM:SIN:MAS", Lemma: "der"}, "Frau"))
	require.Equal(t, []string{"das"}, f.AdaptedDet(DetReading{Token: "der", POS: "ART:DEF:NOM:SIN:NEU", Lemma: "der"}, "Kind"))
	require.Equal(t, []string{"ein"}, f.AdaptedDet(DetReading{Token: "eine", POS: "ART:IND:NOM:SIN:FEM", Lemma: "ein"}, "Plan"))
	require.Equal(t, []string{"einen"}, f.AdaptedDet(DetReading{Token: "eine", POS: "ART:IND:AKK:SIN:FEM", Lemma: "ein"}, "Plan"))
	require.Equal(t, []string{"eines"}, f.AdaptedDet(DetReading{Token: "einer", POS: "ART:IND:GEN:SIN:FEM", Lemma: "ein"}, "Plan"))
	require.Equal(t, []string{"einem"}, f.AdaptedDet(DetReading{Token: "einer", POS: "ART:IND:DAT:SIN:FEM", Lemma: "ein"}, "Plan"))
}

// detAdjNoun is disabled in Java (&& false); soft SuggestWithDetAdj path removed.
func TestAdaptSuggestionFilter_DetAdjBranchDisabled(t *testing.T) {
	// Document: Java accepts only det+noun until detAdj is uncommented upstream.
	// Go does not invent weak-adj synthesis.
	f := withTestHooks(NewAdaptSuggestionFilter())
	require.NotNil(t, f)
}

// Twin of AdaptSuggestionFilterTest.testAcceptRuleMatchDevTest (Java @Ignore for development).
func TestAdaptSuggestionFilter_AcceptRuleMatchDevTest(t *testing.T) {
	t.Skip("Java @Ignore(\"for development\")")
}

// Twin of AdaptSuggestionFilterTest.testAcceptRuleMatchWithDetAdj (Java @Ignore WIP).
func TestAdaptSuggestionFilter_AcceptRuleMatchWithDetAdj(t *testing.T) {
	// Java detAdjNoun branch disabled (&& false); twin is fail-closed / skip.
	t.Skip("Java @Ignore(\"WIP\") — detAdjNoun path not active upstream")
}

// Twin of AdaptSuggestionFilterTest.testdAdaptedDetAdj (Java method name has lowercase d).
func TestAdaptSuggestionFilter_DAdaptedDetAdj(t *testing.T) {
	t.Skip("Java @Ignore(\"WIP\") — getAdaptedDetAdj not active")
}
