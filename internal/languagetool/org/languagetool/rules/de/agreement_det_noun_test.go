package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func atrWithPOS(token, pos, lemma string) *languagetool.AnalyzedTokenReadings {
	var p, l *string
	if pos != "" {
		pp := pos
		p = &pp
	}
	if lemma != "" {
		ll := lemma
		l = &ll
	}
	return languagetool.NewAnalyzedTokenReadingsAt(languagetool.NewAnalyzedToken(token, p, l), 0)
}

func TestGetAgreementCategories_DetAndNoun(t *testing.T) {
	det := atrWithPOS("die", "ART:DEF:NOM:SIN:FEM", "die")
	cats := GetAgreementCategories(det, nil, false)
	require.Contains(t, cats, "NOM/SIN/FEM/DEFINITE")

	noun := atrWithPOS("Haus", "SUB:NOM:SIN:NEU", "Haus")
	ncats := GetAgreementCategories(noun, nil, false)
	require.Contains(t, ncats, "NOM/SIN/NEU/DEFINITE")
	require.False(t, CategoriesIntersect(cats, ncats), "die(FEM) vs Haus(NEU) must not agree")
}

func TestAgreementRule_DetNounMismatch(t *testing.T) {
	toks := []*languagetool.AnalyzedTokenReadings{
		atrWithPOS("die", "ART:DEF:NOM:SIN:FEM", "die"),
		atrWithPOS("Haus", "SUB:NOM:SIN:NEU", "Haus"),
	}
	sent := languagetool.NewAnalyzedSentence(toks)
	r := NewAgreementRule(nil)
	ms := r.Match(sent)
	require.NotEmpty(t, ms)
	require.Equal(t, "Kongruenz von Nominalphrasen (unvollständig!), z.B. 'mein kleiner (kleines) Haus'", r.GetDescription())
	require.Equal(t, "https://languagetool.org/insights/de/beitrag/deklination/", r.GetURL())
	require.Greater(t, r.EstimateContextForSureMatch(), 0)
}

func TestAgreementRule_DetNounOK(t *testing.T) {
	toks := []*languagetool.AnalyzedTokenReadings{
		atrWithPOS("das", "ART:DEF:NOM:SIN:NEU", "das"),
		atrWithPOS("Haus", "SUB:NOM:SIN:NEU", "Haus"),
	}
	sent := languagetool.NewAnalyzedSentence(toks)
	r := NewAgreementRule(nil)
	ms := r.Match(sent)
	require.Empty(t, ms)
}

func TestSuggestionFilter_WithAgreementRule(t *testing.T) {
	// FilterSuggestions path: template "Das ist {}." — open compound still fires on "Original Mail"
	// DET-NOUN needs tags; AnalyzePlain untagged won't mismatch.
	// Unit: direct Match with tags is covered above.
	f := WireAdaptSuggestionFilter()
	require.NotNil(t, f.FilterSuggestions)
	// Filter keeps strings that don't re-trigger AgreementRule
	out := f.FilterSuggestions([]string{"ok phrase"}, "Das ist {}.")
	require.Equal(t, []string{"ok phrase"}, out)
}
