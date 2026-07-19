package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestStyleTooOftenUsedNounRule(t *testing.T) {
	rule := NewStyleTooOftenUsedNounRule(nil)
	// Haus ×2 with SUB: + lemma (Java isToCountedWord SUB:)
	s1 := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Er", "PRO:PER:NOM:SIN:MAS", "er"),
		atrWithPOS("sah", "VER:3:SIN:PRT:NON", "sehen"),
		atrWithPOS("das", "ART:DEF:AKK:SIN:NEU", "der"),
		atrWithPOS("Haus", "SUB:AKK:SIN:NEU", "Haus"),
		atrWithPOS(".", "PKT", "."),
	))
	s2 := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Dann", "ADV", "dann"),
		atrWithPOS("kaufte", "VER:3:SIN:PRT:SFT", "kaufen"),
		atrWithPOS("er", "PRO:PER:NOM:SIN:MAS", "er"),
		atrWithPOS("das", "ART:DEF:AKK:SIN:NEU", "der"),
		atrWithPOS("Haus", "SUB:AKK:SIN:NEU", "Haus"),
		atrWithPOS(".", "PKT", "."),
	))
	matches := rule.MatchList([]*languagetool.AnalyzedSentence{s1, s2})
	require.GreaterOrEqual(t, len(matches), 2)

	// different nouns
	s3 := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Dann", "ADV", "dann"),
		atrWithPOS("kaufte", "VER:3:SIN:PRT:SFT", "kaufen"),
		atrWithPOS("er", "PRO:PER:NOM:SIN:MAS", "er"),
		atrWithPOS("die", "ART:DEF:AKK:SIN:FEM", "der"),
		atrWithPOS("Villa", "SUB:AKK:SIN:FEM", "Villa"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 0, len(rule.MatchList([]*languagetool.AnalyzedSentence{s1, s3})))

	// untagged must not invent
	plain := languagetool.SplitAndAnalyze("Er sah das Haus am See. Dann kaufte er das Haus in der Stadt.")
	require.Equal(t, 0, len(rule.MatchList(plain)))
}
