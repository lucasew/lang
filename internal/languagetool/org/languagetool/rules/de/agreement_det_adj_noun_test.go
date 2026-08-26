package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestAgreementRule_DetAdjNounMismatch(t *testing.T) {
	// die riesigen Tisch — DET FEM + ADJ PLU + SUB SIN MAS
	toks := []*languagetool.AnalyzedTokenReadings{
		atrWithPOS("die", "ART:DEF:NOM:SIN:FEM", "die"),
		atrWithPOS("riesigen", "ADJ:AKK:SIN:MAS:GRU:DEF", "riesig"),
		atrWithPOS("Tisch", "SUB:NOM:SIN:MAS", "Tisch"),
	}
	// riesigen DEF SIN MAS + die FEM + Tisch MAS → may or may not intersect depending on categories
	// clearer mismatch: der + große + Haus (NEU)
	toks = []*languagetool.AnalyzedTokenReadings{
		atrWithPOS("der", "ART:DEF:NOM:SIN:MAS", "der"),
		atrWithPOS("große", "ADJ:NOM:SIN:FEM:GRU:DEF", "groß"),
		atrWithPOS("Haus", "SUB:NOM:SIN:NEU", "Haus"),
	}
	sent := languagetool.NewAnalyzedSentence(toks)
	r := NewAgreementRule(nil)
	ms := r.Match(sent)
	require.NotEmpty(t, ms, "der große Haus should mismatch")
}

func TestAgreementRule_DetAdjNounOK(t *testing.T) {
	toks := []*languagetool.AnalyzedTokenReadings{
		atrWithPOS("das", "ART:DEF:NOM:SIN:NEU", "das"),
		atrWithPOS("große", "ADJ:NOM:SIN:NEU:GRU:DEF", "groß"),
		atrWithPOS("Haus", "SUB:NOM:SIN:NEU", "Haus"),
	}
	sent := languagetool.NewAnalyzedSentence(toks)
	r := NewAgreementRule(nil)
	ms := r.Match(sent)
	// filter open-compound only; morphological should not fire
	for _, m := range ms {
		require.NotEqual(t, agreementShort, m.ShortMessage)
	}
}

func TestAgreementRule_DetAdjAdjNounMismatch(t *testing.T) {
	toks := []*languagetool.AnalyzedTokenReadings{
		atrWithPOS("die", "ART:DEF:NOM:SIN:FEM", "die"),
		atrWithPOS("große", "ADJ:NOM:SIN:FEM:GRU:DEF", "groß"),
		atrWithPOS("riesige", "ADJ:NOM:SIN:FEM:GRU:DEF", "riesig"),
		atrWithPOS("Tisch", "SUB:NOM:SIN:MAS", "Tisch"),
	}
	sent := languagetool.NewAnalyzedSentence(toks)
	r := NewAgreementRule(nil)
	ms := r.Match(sent)
	require.NotEmpty(t, ms)
}

func TestRetainCommonCategories3(t *testing.T) {
	det := atrWithPOS("das", "ART:DEF:NOM:SIN:NEU", "das")
	adj := atrWithPOS("große", "ADJ:NOM:SIN:NEU:GRU:DEF", "groß")
	noun := atrWithPOS("Haus", "SUB:NOM:SIN:NEU", "Haus")
	common := retainCommonCategories3(det, adj, noun)
	require.NotEmpty(t, common)
}
