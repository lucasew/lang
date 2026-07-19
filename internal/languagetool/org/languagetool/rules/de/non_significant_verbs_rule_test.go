package de

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestNonSignificantVerbsRule(t *testing.T) {
	rule := NewNonSignificantVerbsRule(nil)
	// Java: hasLemma("machen") — inject lemma
	machte := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Er", "PRO:PER:NOM:SIN:MAS", "er"),
		atrWithPOS("machte", "VER:3:SIN:PRT:SFT", "machen"),
		atrWithPOS("einen", "ART:IND:AKK:SIN:MAS", "ein"),
		atrWithPOS("Kuchen", "SUB:AKK:SIN:MAS", "Kuchen"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 1, len(rule.Match(machte)))
	// Angst exception with machen
	angst := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Das", "PRO:DEM:NOM:SIN:NEU", "das"),
		atrWithPOS("macht", "VER:3:SIN:PRS:SFT", "machen"),
		atrWithPOS("mir", "PRO:PER:DAT:SIN:MAS", "ich"),
		atrWithPOS("Angst", "SUB:AKK:SIN:FEM", "Angst"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 0, len(rule.Match(angst)))
	// untagged must not invent
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Er machte einen Kuchen."))))
}
