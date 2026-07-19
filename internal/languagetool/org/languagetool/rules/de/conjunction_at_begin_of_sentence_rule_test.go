package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestConjunctionAtBeginOfSentenceRule(t *testing.T) {
	rule := NewConjunctionAtBeginOfSentenceRule(nil)
	// Java: hasPosTagStartingWith("KON") only
	und := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Und", "KON:NEB", "und"),
		atrWithPOS("dann", "ADV", "dann"),
		atrWithPOS("ging", "VER:3:SIN:PRT:NON", "gehen"),
		atrWithPOS("er", "PRO:PER:NOM:SIN:MAS", "er"),
		atrWithPOS("weg", "ADV", "weg"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 1, len(rule.Match(und)))
	// "Wie" exception (even with KON)
	wie := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Wie", "KON:UNT", "wie"),
		atrWithPOS("geht", "VER:3:SIN:PRS:NON", "gehen"),
		atrWithPOS("es", "PRO:PER:NOM:SIN:NEU", "es"),
		atrWithPOS("dir", "PRO:PER:DAT:SIN:MAS", "du"),
		atrWithPOS("?", "PKT", "?"),
	))
	require.Equal(t, 0, len(rule.Match(wie)))
	// not a conjunction
	er := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Er", "PRO:PER:NOM:SIN:MAS", "er"),
		atrWithPOS("ging", "VER:3:SIN:PRT:NON", "gehen"),
		atrWithPOS("dann", "ADV", "dann"),
		atrWithPOS("weg", "ADV", "weg"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 0, len(rule.Match(er)))
	// untagged must not invent
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Und dann ging er weg."))))
}
