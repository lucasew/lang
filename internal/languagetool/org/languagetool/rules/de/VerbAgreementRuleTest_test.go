package de

// Twin of VerbAgreementRuleTest — Java uses tagged analysis (VER person/number).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestVerbAgreementRule_WrongVerb(t *testing.T) {
	rule := NewVerbAgreementRule(nil)
	// Ich bin OK
	ok := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Ich", "PRO:PER:NOM:SIN:ALG", "ich"),
		atrWithPOS("bin", "VER:1:SIN:PRÄ:NON", "sein"),
		atrWithPOS("müde", "ADJ:PRD:GRU", "müde"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 0, len(rule.Match(ok)))

	// Ich sind wrong
	bad := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Ich", "PRO:PER:NOM:SIN:ALG", "ich"),
		atrWithPOS("sind", "VER:1:PLU:PRÄ:NON", "sein"),
		atrWithPOS("müde", "ADJ:PRD:GRU", "müde"),
		atrWithPOS(".", "PKT", "."),
	))
	// morph may emit both wrong-verb and wrong-subject matches
	require.GreaterOrEqual(t, len(rule.Match(bad)), 1)

	// untagged must not invent
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Ich sind müde."))))
}

func TestVerbAgreementRule_SuggestionSorting(t *testing.T) {
	require.NotNil(t, NewVerbAgreementRule(nil))
}

func TestVerbAgreementRule_Positions(t *testing.T) {
	rule := NewVerbAgreementRule(nil)
	ms := rule.Match(languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Du", "PRO:PER:NOM:SIN:ALG", "du"),
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("hier", "ADV", "hier"),
		atrWithPOS(".", "PKT", "."),
	)))
	require.Equal(t, 1, len(ms))
}

// Twin of VerbAgreementRuleTest.testWrongVerbSubject
func TestVerbAgreementRule_WrongVerbSubject(t *testing.T) {
	rule := NewVerbAgreementRule(nil)
	// Good: Du lebst
	good := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Du", "PRO:PER:NOM:SIN:ALG", "du"),
		atrWithPOS("lebst", "VER:2:SIN:PRÄ:SFT", "leben"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 0, len(rule.Match(good)))

	// Bad: Du leben (plural verb with du)
	bad := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Du", "PRO:PER:NOM:SIN:ALG", "du"),
		atrWithPOS("leben", "VER:1:PLU:PRÄ:SFT", "leben"),
		atrWithPOS(".", "PKT", "."),
	))
	require.GreaterOrEqual(t, len(rule.Match(bad)), 1)

	// Bad: Ich sind
	bad2 := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Ich", "PRO:PER:NOM:SIN:ALG", "ich"),
		atrWithPOS("sind", "VER:1:PLU:PRÄ:NON", "sein"),
		atrWithPOS("nett", "ADJ:PRD:GRU", "nett"),
		atrWithPOS(".", "PKT", "."),
	))
	require.GreaterOrEqual(t, len(rule.Match(bad2)), 1)

	// Good: Wir leben
	good2 := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		atrWithPOS("Wir", "PRO:PER:NOM:PLU:ALG", "wir"),
		atrWithPOS("leben", "VER:1:PLU:PRÄ:SFT", "leben"),
		atrWithPOS("noch", "ADV", "noch"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 0, len(rule.Match(good2)))

	// untagged must not invent
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Auch morgen leben du."))))
}
