package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/AdaptSuggestionFilterTest.java
import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Port of AdaptSuggestionFilterTest.testAcceptRuleMatchDevTest
func TestAdaptSuggestionFilter_AcceptRuleMatchDevTest(t *testing.T) {
	t.Skip("Java @Ignore")
}

// Port of AdaptSuggestionFilterTest.testAcceptRuleMatchWithDet
func TestAdaptSuggestionFilter_AcceptRuleMatchWithDet(t *testing.T) {
	// Surface twin of a subset of Java testAcceptRuleMatchWithDet (no full JLanguageTool pipeline).
	f := NewAdaptSuggestionFilter()
	// MAS (der Plan) after FEM det "die":
	require.Equal(t, []string{"der Plan"}, f.SuggestWithDet("die", "ART:DEF:NOM:SIN:FEM", "der", []string{"Plan"}))
	require.Equal(t, []string{"ein Plan"}, f.SuggestWithDet("eine", "ART:IND:NOM:SIN:FEM", "ein", []string{"Plan"}))
	require.Equal(t, []string{"mein Plan"}, f.SuggestWithDet("meine", "PRO:POS:NOM:SIN:FEM", "mein", []string{"Plan"}))
	// FEM (die Idee):
	require.Equal(t, []string{"die Idee"}, f.SuggestWithDet("die", "ART:DEF:NOM:SIN:FEM", "der", []string{"Idee"}))
	require.Equal(t, []string{"meine Idee"}, f.SuggestWithDet("meine", "PRO:POS:NOM:SIN:FEM", "mein", []string{"Idee"}))
	// NEU (das Verfahren):
	require.Equal(t, []string{"das Verfahren"}, f.SuggestWithDet("die", "ART:DEF:NOM:SIN:FEM", "der", []string{"Verfahren"}))
	require.Equal(t, []string{"mein Verfahren"}, f.SuggestWithDet("meine", "PRO:POS:NOM:SIN:FEM", "mein", []string{"Verfahren"}))
}

// Port of AdaptSuggestionFilterTest.testAcceptRuleMatchWithDetAdj
func TestAdaptSuggestionFilter_AcceptRuleMatchWithDetAdj(t *testing.T) {
	f := NewAdaptSuggestionFilter()
	// die + schöne + Plan → der schöne Plan (masc)
	got := f.SuggestWithDetAdj("die", "ART:DEF:NOM:SIN:FEM", "der", "schöne", []string{"Plan"})
	require.Contains(t, got, "der schöne Plan")
	// meine + gute + Idee
	got2 := f.SuggestWithDetAdj("meine", "PRO:POS:NOM:SIN:FEM", "mein", "gute", []string{"Idee"})
	require.Contains(t, got2, "meine gute Idee")
}

// Port of AdaptSuggestionFilterTest.testAdaptedDet
func TestAdaptSuggestionFilter_AdaptedDet_Twin(t *testing.T) {
	f := NewAdaptSuggestionFilter()
	require.Equal(t, []string{"der"}, f.AdaptedDet(DetReading{Token: "die", POS: "ART:DEF:NOM:SIN:FEM", Lemma: "der"}, "Mann"))
	require.Equal(t, []string{"die"}, f.AdaptedDet(DetReading{Token: "der", POS: "ART:DEF:NOM:SIN:MAS", Lemma: "der"}, "Frau"))
	require.Equal(t, []string{"das"}, f.AdaptedDet(DetReading{Token: "der", POS: "ART:DEF:NOM:SIN:NEU", Lemma: "der"}, "Kind"))
	require.Equal(t, []string{"ein"}, f.AdaptedDet(DetReading{Token: "eine", POS: "ART:IND:NOM:SIN:FEM", Lemma: "ein"}, "Plan"))
	require.Equal(t, []string{"einen"}, f.AdaptedDet(DetReading{Token: "eine", POS: "ART:IND:AKK:SIN:FEM", Lemma: "ein"}, "Plan"))
	require.Equal(t, []string{"eines"}, f.AdaptedDet(DetReading{Token: "einer", POS: "ART:IND:GEN:SIN:FEM", Lemma: "ein"}, "Plan"))
	require.Equal(t, []string{"einem"}, f.AdaptedDet(DetReading{Token: "einer", POS: "ART:IND:DAT:SIN:FEM", Lemma: "ein"}, "Plan"))
}

// Port of AdaptSuggestionFilterTest.testdAdaptedDetAdj
func TestAdaptSuggestionFilter_DAdaptedDetAdj(t *testing.T) {
	t.Skip("Java @Ignore / needs adjective synthesizer")
}
