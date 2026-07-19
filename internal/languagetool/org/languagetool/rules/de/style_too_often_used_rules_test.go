package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestStyleTooOftenUsedVerbRule(t *testing.T) {
	rule := NewStyleTooOftenUsedVerbRule(nil)
	// laufen ×2 with VER: + lemma
	s1 := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Sie", "PRO:PER:NOM:PLU:MAS", "sie"),
		atrWithPOS("laufen", "VER:3:PLU:PRS:NON", "laufen"),
		atrWithPOS("schnell", "ADV", "schnell"),
		atrWithPOS(".", "PKT", "."),
	))
	s2 := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Dann", "ADV", "dann"),
		atrWithPOS("laufen", "VER:3:PLU:PRS:NON", "laufen"),
		atrWithPOS("sie", "PRO:PER:NOM:PLU:MAS", "sie"),
		atrWithPOS("weiter", "ADV", "weiter"),
		atrWithPOS(".", "PKT", "."),
	))
	require.GreaterOrEqual(t, len(rule.MatchList([]*languagetool.AnalyzedSentence{s1, s2})), 2)
	// untagged no invent
	require.Equal(t, 0, len(rule.MatchList(languagetool.SplitAndAnalyze("Sie laufen schnell. Dann laufen sie weiter."))))
}

func TestStyleTooOftenUsedAdjectiveRule(t *testing.T) {
	rule := NewStyleTooOftenUsedAdjectiveRule(nil)
	s1 := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Ein", "ART:IND:NOM:SIN:NEU", "ein"),
		atrWithPOS("schönes", "ADJ:NOM:SIN:NEU:GRU:IND", "schön"),
		atrWithPOS("Auto", "SUB:NOM:SIN:NEU", "Auto"),
		atrWithPOS(".", "PKT", "."),
	))
	s2 := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Noch", "ADV", "noch"),
		atrWithPOS("ein", "ART:IND:NOM:SIN:NEU", "ein"),
		atrWithPOS("schönes", "ADJ:NOM:SIN:NEU:GRU:IND", "schön"),
		atrWithPOS("Haus", "SUB:NOM:SIN:NEU", "Haus"),
		atrWithPOS(".", "PKT", "."),
	))
	require.GreaterOrEqual(t, len(rule.MatchList([]*languagetool.AnalyzedSentence{s1, s2})), 2)
	require.Equal(t, 0, len(rule.MatchList(languagetool.SplitAndAnalyze("Ein schönes Auto. Noch ein schönes Haus."))))
}
