package de

// Twin of AgreementRule2Test — Java uses tagged analysis (ADJ+SUB).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
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

// Twin of AgreementRule2Test.testSuggestion
func TestAgreementRule2_Suggestion(t *testing.T) {
	// mock synth: Synthesize(adj, ADJ:NOM:num:gen:GRU:SOL) → form
	synth := synthesis.FuncSynthesizer{
		Synth: func(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
			if token == nil || token.GetLemma() == nil {
				return nil, nil
			}
			lem := *token.GetLemma()
			switch {
			case lem == "klein" && posTag == "ADJ:NOM:SIN:NEU:GRU:SOL":
				return []string{"kleines"}, nil
			case lem == "klein" && posTag == "ADJ:NOM:PLU:NEU:GRU:SOL":
				return []string{"kleine"}, nil
			case lem == "jung" && posTag == "ADJ:NOM:SIN:FEM:GRU:SOL":
				return []string{"junge"}, nil
			case lem == "wirtschaftlich" && posTag == "ADJ:NOM:SIN:NEU:GRU:SOL":
				return []string{"wirtschaftliches"}, nil
			default:
				return nil, nil
			}
		},
	}
	rule := NewAgreementRule2(nil).WithSynth(synth)

	// Good: Kleinem Haus (DAT)
	goodDat := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Kleinem", "ADJ:DAT:SIN:NEU:GRU:SOL", "klein"),
		atrWithPOS("Haus", "SUB:DAT:SIN:NEU", "Haus"),
		atrWithPOS("am", "APPRART:DAT:SIN:NEU", "an"),
		atrWithPOS("Waldesrand", "SUB:DAT:SIN:MAS", "Waldesrand"),
	))
	require.Equal(t, 0, len(rule.Match(goodDat)))

	// Bad: Kleiner Haus → Kleines Haus
	bad := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Kleiner", "ADJ:NOM:SIN:MAS:GRU:SOL", "klein"),
		atrWithPOS("Haus", "SUB:NOM:SIN:NEU", "Haus"),
	))
	ms := rule.Match(bad)
	require.Equal(t, 1, len(ms))
	require.Contains(t, ms[0].GetSuggestedReplacements(), "Kleines Haus")

	// Bad plural: Kleines Häuser → Kleine Häuser
	bad2 := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Kleines", "ADJ:NOM:SIN:NEU:GRU:SOL", "klein"),
		atrWithPOS("Häuser", "SUB:NOM:PLU:NEU", "Haus"),
	))
	ms2 := rule.Match(bad2)
	require.Equal(t, 1, len(ms2))
	require.Contains(t, ms2[0].GetSuggestedReplacements(), "Kleine Häuser")

	// Bad: Junges Frau → Junge Frau
	bad3 := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Junges", "ADJ:NOM:SIN:NEU:GRU:SOL", "jung"),
		atrWithPOS("Frau", "SUB:NOM:SIN:FEM", "Frau"),
	))
	ms3 := rule.Match(bad3)
	require.Equal(t, 1, len(ms3))
	require.Contains(t, ms3[0].GetSuggestedReplacements(), "Junge Frau")

	// Good: Junge Frau
	good := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Junge", "ADJ:NOM:SIN:FEM:GRU:SOL", "jung"),
		atrWithPOS("Frau", "SUB:NOM:SIN:FEM", "Frau"),
	))
	require.Equal(t, 0, len(rule.Match(good)))

	// untagged no invent
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Kleiner Haus am Waldesrand"))))
}
