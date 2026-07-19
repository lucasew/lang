package de

// Twin of GermanStyleRepeatedWordRuleTest — Java uses tagged analysis (ADJ/SUB/VER).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestGermanStyleRepeatedWordRule_Rule(t *testing.T) {
	rule := NewGermanStyleRepeatedWordRule(nil)
	// "großen" ADJ ×2 across sentences
	s1 := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Der", "ART:DEF:NOM:SIN:MAS", "der"),
		atrWithPOS("alte", "ADJ:NOM:SIN:MAS:GRU:DEF", "alt"),
		atrWithPOS("Mann", "SUB:NOM:SIN:MAS", "Mann"),
		atrWithPOS("wohnte", "VER:3:SIN:PRT:SFT", "wohnen"),
		atrWithPOS("in", "PRP:DAT", "in"),
		atrWithPOS("einem", "ART:IND:DAT:SIN:NEU", "ein"),
		atrWithPOS("großen", "ADJ:DAT:SIN:NEU:GRU:IND", "groß"),
		atrWithPOS("Haus", "SUB:DAT:SIN:NEU", "Haus"),
		atrWithPOS(".", "PKT", "."),
	))
	s2 := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Es", "PRO:PER:NOM:SIN:NEU", "es"),
		atrWithPOS("stand", "VER:3:SIN:PRT:NON", "stehen"),
		atrWithPOS("in", "PRP:DAT", "in"),
		atrWithPOS("einem", "ART:IND:DAT:SIN:NEU", "ein"),
		atrWithPOS("großen", "ADJ:DAT:SIN:NEU:GRU:IND", "groß"),
		atrWithPOS("Garten", "SUB:DAT:SIN:MAS", "Garten"),
		atrWithPOS(".", "PKT", "."),
	))
	// both "großen" flagged (same sentence distance)
	require.Equal(t, 2, len(rule.MatchList([]*languagetool.AnalyzedSentence{s1, s2})))

	// different adjectives
	s2b := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Es", "PRO:PER:NOM:SIN:NEU", "es"),
		atrWithPOS("stand", "VER:3:SIN:PRT:NON", "stehen"),
		atrWithPOS("in", "PRP:DAT", "in"),
		atrWithPOS("einem", "ART:IND:DAT:SIN:NEU", "ein"),
		atrWithPOS("weitläufigen", "ADJ:DAT:SIN:NEU:GRU:IND", "weitläufig"),
		atrWithPOS("Garten", "SUB:DAT:SIN:MAS", "Garten"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{s1, s2b})))

	// Java isUnknownWord: isPosTagUnknown only (not invent !isTagged).
	unk := languagetool.AnalyzePlain("Blahxyz").GetTokensWithoutWhitespace()
	var blah *languagetool.AnalyzedTokenReadings
	for _, tok := range unk {
		if tok != nil && tok.GetToken() == "Blahxyz" {
			blah = tok
			break
		}
	}
	require.NotNil(t, blah)
	require.True(t, blah.IsPosTagUnknown())
	require.True(t, isUnknownWordStyle(blah))
	// Tagged SUB is not "unknown word" for this gate.
	require.False(t, isUnknownWordStyle(atrWithPOS("Haus", "SUB:NOM:SIN:NEU", "Haus")))
}
