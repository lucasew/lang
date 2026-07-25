package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Temporary home for PassiveSentenceRule smoke tests moved out of SentenceWithManRule sector.
// Full PassiveSentenceRule sector will raise this to Modal-level twin coverage later.
func TestPassiveSentenceRule(t *testing.T) {
	rule := NewPassiveSentenceRuleWithMinPercent(nil, 0)
	// Java: lemma werden + VER:PA2
	hit := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Das", "ART:DEF:NOM:SIN:NEU", "der"),
		atrWithPOS("Haus", "SUB:NOM:SIN:NEU", "Haus"),
		atrWithPOS("wird", "VER:AUX:3:SIN:PRS:SFT", "werden"),
		atrWithPOS("gebaut", "VER:PA2:NON", "bauen"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 1, len(rule.Match(hit)))
	active := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Er", "PRO:PER:NOM:SIN:MAS", "er"),
		atrWithPOS("baut", "VER:3:SIN:PRS:NON", "bauen"),
		atrWithPOS("das", "ART:DEF:AKK:SIN:NEU", "der"),
		atrWithPOS("Haus", "SUB:AKK:SIN:NEU", "Haus"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 0, len(rule.Match(active)))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Das Haus wird gebaut."))))
}
