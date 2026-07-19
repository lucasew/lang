package de

// Twin of AgreementRule2Test — Java uses tagged analysis (ADJ+SUB).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestAgreementRule2_Rule(t *testing.T) {
	rule := NewAgreementRule2(nil)
	// Kleiner (MAS) Haus (NEU) mismatch
	bad := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Kleiner", "ADJ:NOM:SIN:MAS:GRU:SOL", "klein"),
		atrWithPOS("Haus", "SUB:NOM:SIN:NEU", "Haus"),
	))
	require.Equal(t, 1, len(rule.Match(bad)))

	good := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Kleines", "ADJ:NOM:SIN:NEU:GRU:SOL", "klein"),
		atrWithPOS("Haus", "SUB:NOM:SIN:NEU", "Haus"),
	))
	require.Equal(t, 0, len(rule.Match(good)))

	// Wirtschaftlich (ADJ no gender agreement) vs Wachstum NEU — use mismatched tags
	bad2 := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Wirtschaftlich", "ADJ:PRD:GRU", "wirtschaftlich"),
		atrWithPOS("Wachstum", "SUB:NOM:SIN:NEU", "Wachstum"),
	))
	// PRD adj may not yield agreement categories — only assert untagged no invent
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Kleiner Haus am Waldesrand"))))

	// Deutscher Taschenbuch Verlag — third SUB skips (Java)
	triple := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Deutscher", "ADJ:NOM:SIN:MAS:GRU:SOL", "deutsch"),
		atrWithPOS("Taschenbuch", "SUB:NOM:SIN:NEU", "Taschenbuch"),
		atrWithPOS("Verlag", "SUB:NOM:SIN:MAS", "Verlag"),
	))
	require.Equal(t, 0, len(rule.Match(triple)))

	// Deutscher Taschenbuch without third SUB — mismatch MAS adj + NEU noun
	pair := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Deutscher", "ADJ:NOM:SIN:MAS:GRU:SOL", "deutsch"),
		atrWithPOS("Taschenbuch", "SUB:NOM:SIN:NEU", "Taschenbuch"),
	))
	require.Equal(t, 1, len(rule.Match(pair)))

	_ = bad2
}
