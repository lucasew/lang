package de

// Twin of PassiveSentenceRule (Java: hasLemma("werden") + VER:PA2 — no surface invent).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestPassiveSentenceRule_Match(t *testing.T) {
	r := NewPassiveSentenceRuleWithMinPercent(nil, 0) // show-all for twin morph
	hit := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Das", "ART:DEF:NOM:SIN:NEU", "der"),
		atrWithPOS("Haus", "SUB:NOM:SIN:NEU", "Haus"),
		atrWithPOS("wird", "VER:AUX:3:SIN:PRS:SFT", "werden"),
		atrWithPOS("gebaut", "VER:PA2:NON", "bauen"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 1, len(r.Match(hit)))
	hit2 := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Die", "ART:DEF:NOM:SIN:FEM", "der"),
		atrWithPOS("Tür", "SUB:NOM:SIN:FEM", "Tür"),
		atrWithPOS("wurde", "VER:AUX:3:SIN:PRT:SFT", "werden"),
		atrWithPOS("geöffnet", "VER:PA2:SFT", "öffnen"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 1, len(r.Match(hit2)))
	active := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Er", "PRO:PER:NOM:SIN:MAS", "er"),
		atrWithPOS("baut", "VER:3:SIN:PRS:NON", "bauen"),
		atrWithPOS("das", "ART:DEF:AKK:SIN:NEU", "der"),
		atrWithPOS("Haus", "SUB:AKK:SIN:NEU", "Haus"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 0, len(r.Match(active)))
	// untagged must not invent
	require.Equal(t, 0, len(r.Match(languagetool.AnalyzePlain("Das Haus wird gebaut."))))
}
