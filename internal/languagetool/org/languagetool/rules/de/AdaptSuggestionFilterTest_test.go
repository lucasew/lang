package de

// Twin of languagetool-language-modules/de/src/test/java/org/languagetool/rules/de/AdaptSuggestionFilterTest.java
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Port of AdaptSuggestionFilterTest.testAcceptRuleMatchWithDet (surface det path).
func TestAdaptSuggestionFilter_AcceptRuleMatchWithDet(t *testing.T) {
	f := withTestHooks(NewAdaptSuggestionFilter())
	// Single NOM reading → one adapted det (unit path).
	require.Equal(t, []string{"der Plan"}, f.SuggestWithDet("die", "ART:DEF:NOM:SIN:FEM", "der", []string{"Plan"}))
	require.Equal(t, []string{"ein Plan"}, f.SuggestWithDet("eine", "ART:IND:NOM:SIN:FEM", "ein", []string{"Plan"}))
	require.Equal(t, []string{"mein Plan"}, f.SuggestWithDet("meine", "PRO:POS:NOM:SIN:FEM", "mein", []string{"Plan"}))
	require.Equal(t, []string{"die Idee"}, f.SuggestWithDet("die", "ART:DEF:NOM:SIN:FEM", "der", []string{"Idee"}))
	require.Equal(t, []string{"meine Idee"}, f.SuggestWithDet("meine", "PRO:POS:NOM:SIN:FEM", "mein", []string{"Idee"}))
	require.Equal(t, []string{"das Verfahren"}, f.SuggestWithDet("die", "ART:DEF:NOM:SIN:FEM", "der", []string{"Verfahren"}))
	require.Equal(t, []string{"mein Verfahren"}, f.SuggestWithDet("meine", "PRO:POS:NOM:SIN:FEM", "mein", []string{"Verfahren"}))

	// Java Morphy "die" carries NOM+AKK FEM → MAS yields Der + Den (order unique, both present).
	dieATR := atrWithPOS("die", "ART:DEF:NOM:SIN:FEM", "der")
	akk := "ART:DEF:AKK:SIN:FEM"
	lem := "der"
	dieATR.AddReading(languagetool.NewAnalyzedToken("die", &akk, &lem), "")
	got := f.SuggestWithDetFromATR(dieATR, []string{"Plan"})
	require.ElementsMatch(t, []string{"der Plan", "den Plan"}, got)

	// Possessive table (Java Hier steht meine/deine/… Roadmap → Plan / Idee / Verfahren)
	type row struct {
		tok, pos, lemma, noun string
		want                  []string
	}
	for _, tc := range []row{
		{"meine", "PRO:POS:NOM:SIN:FEM", "mein", "Plan", []string{"mein Plan"}},
		{"deine", "PRO:POS:NOM:SIN:FEM", "dein", "Plan", []string{"dein Plan"}},
		{"seine", "PRO:POS:NOM:SIN:FEM", "sein", "Plan", []string{"sein Plan"}},
		{"ihre", "PRO:POS:NOM:SIN:FEM", "ihr", "Plan", []string{"ihr Plan"}},
		{"unsere", "PRO:POS:NOM:SIN:FEM", "unser", "Plan", []string{"unser Plan"}},
		{"eure", "PRO:POS:NOM:SIN:FEM", "euer", "Plan", []string{"euer Plan"}},
		{"meine", "PRO:POS:NOM:SIN:FEM", "mein", "Idee", []string{"meine Idee"}},
		{"deine", "PRO:POS:NOM:SIN:FEM", "dein", "Idee", []string{"deine Idee"}},
		{"seine", "PRO:POS:NOM:SIN:FEM", "sein", "Idee", []string{"seine Idee"}},
		{"ihre", "PRO:POS:NOM:SIN:FEM", "ihr", "Idee", []string{"ihre Idee"}},
		{"unsere", "PRO:POS:NOM:SIN:FEM", "unser", "Idee", []string{"unsere Idee"}},
		{"eure", "PRO:POS:NOM:SIN:FEM", "euer", "Idee", []string{"eure Idee"}},
		{"meine", "PRO:POS:NOM:SIN:FEM", "mein", "Verfahren", []string{"mein Verfahren"}},
		{"deine", "PRO:POS:NOM:SIN:FEM", "dein", "Verfahren", []string{"dein Verfahren"}},
		{"eine", "ART:IND:NOM:SIN:FEM", "ein", "Verfahren", []string{"ein Verfahren"}},
		// AKK FEM det → MAS AKK forms for Plan
		{"meine", "PRO:POS:AKK:SIN:FEM", "mein", "Plan", []string{"meinen Plan"}},
		{"deine", "PRO:POS:AKK:SIN:FEM", "dein", "Plan", []string{"deinen Plan"}},
		{"seine", "PRO:POS:AKK:SIN:FEM", "sein", "Plan", []string{"seinen Plan"}},
		{"ihre", "PRO:POS:AKK:SIN:FEM", "ihr", "Plan", []string{"ihren Plan"}},
		{"unsere", "PRO:POS:AKK:SIN:FEM", "unser", "Plan", []string{"unseren Plan"}},
		{"eure", "PRO:POS:AKK:SIN:FEM", "euer", "Plan", []string{"euren Plan"}},
	} {
		got := f.SuggestWithDet(tc.tok, tc.pos, tc.lemma, []string{tc.noun})
		require.Equal(t, tc.want, got, "tok=%s noun=%s", tc.tok, tc.noun)
	}
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
	// Title case first-char filter (Java Die → Der)
	require.Equal(t, []string{"Der"}, f.AdaptedDet(DetReading{Token: "Die", POS: "ART:DEF:NOM:SIN:FEM", Lemma: "der"}, "Plan"))
	// Possessive first-char filter: do not invent "dein" from lemma mein for "meine"
	require.Equal(t, []string{"mein"}, f.AdaptedDet(DetReading{Token: "meine", POS: "PRO:POS:NOM:SIN:FEM", Lemma: "mein"}, "Plan"))
	require.Equal(t, []string{"dein"}, f.AdaptedDet(DetReading{Token: "deine", POS: "PRO:POS:NOM:SIN:FEM", Lemma: "dein"}, "Plan"))
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
