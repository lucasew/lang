package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestStyleTooOftenUsedNounRule_JavaDefaults(t *testing.T) {
	rule := NewStyleTooOftenUsedNounRule(nil)
	require.Equal(t, 5, rule.MinPercent)
	require.Equal(t, 100, rule.MinWordCount)
	// Short texts: numWords < MIN_WORD_COUNT → empty (Java)
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
	require.Empty(t, rule.MatchList([]*languagetool.AnalyzedSentence{s1, s2}))
}

func TestStyleTooOftenUsedNounRule_CountGateOff(t *testing.T) {
	// Force MinWordCount 0 to exercise percent threshold (Java uses 100 in production).
	rule := NewStyleTooOftenUsedNounRule(nil)
	rule.MinWordCount = 0
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
	// Haus is 100% of nouns → both hits
	require.GreaterOrEqual(t, len(rule.MatchList([]*languagetool.AnalyzedSentence{s1, s2})), 2)

	// Haus + Villa each 50% ≥ 5 → both lemmas flag (Java percent >= minPercent)
	s3 := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Dann", "ADV", "dann"),
		atrWithPOS("kaufte", "VER:3:SIN:PRT:SFT", "kaufen"),
		atrWithPOS("er", "PRO:PER:NOM:SIN:MAS", "er"),
		atrWithPOS("die", "ART:DEF:AKK:SIN:FEM", "der"),
		atrWithPOS("Villa", "SUB:AKK:SIN:FEM", "Villa"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 2, len(rule.MatchList([]*languagetool.AnalyzedSentence{s1, s3})))

	// Raise threshold above 50%: neither lemma flags
	rule.MinPercent = 51
	require.Empty(t, rule.MatchList([]*languagetool.AnalyzedSentence{s1, s3}))

	// untagged must not invent
	rule.MinPercent = 5
	plain := languagetool.SplitAndAnalyze("Er sah das Haus am See. Dann kaufte er das Haus in der Stadt.")
	require.Equal(t, 0, len(rule.MatchList(plain)))
}
