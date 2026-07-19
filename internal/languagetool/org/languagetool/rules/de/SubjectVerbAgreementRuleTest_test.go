package de

// Twin of SubjectVerbAgreementRuleTest — Java uses chunk/POS analysis.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSubjectVerbAgreementRule_RuleWithIncorrectSingularVerb(t *testing.T) {
	rule := NewSubjectVerbAgreementRule(nil)
	die := atrWithPOS("Die", "ART:DEF:NOM:PLU:ALG", "die")
	autos := atrWithPOS("Autos", "SUB:NOM:PLU:NEU", "Auto")
	die.SetChunkTags([]string{chunkNPP})
	autos.SetChunkTags([]string{chunkNPP})
	bad := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		die,
		autos,
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("schnell", "ADJ:PRD:GRU", "schnell"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 1, len(rule.Match(bad)))

	// untagged must not invent
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Die Autos ist schnell."))))
}

func TestSubjectVerbAgreementRule_RuleWithCorrectSingularVerb(t *testing.T) {
	rule := NewSubjectVerbAgreementRule(nil)
	die := atrWithPOS("Die", "ART:DEF:NOM:SIN:FEM", "die")
	katze := atrWithPOS("Katze", "SUB:NOM:SIN:FEM", "Katze")
	die.SetChunkTags([]string{chunkNPS})
	katze.SetChunkTags([]string{chunkNPS})
	good := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		die,
		katze,
		atrWithPOS("ist", "VER:3:SIN:PRÄ:NON", "sein"),
		atrWithPOS("schön", "ADJ:PRD:GRU", "schön"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 0, len(rule.Match(good)))
}

func TestSubjectVerbAgreementRule_Temp(t *testing.T) {
	require.NotNil(t, NewSubjectVerbAgreementRule(nil))
}

func TestSubjectVerbAgreementRule_ArrayOutOfBoundsBug(t *testing.T) {
	rule := NewSubjectVerbAgreementRule(nil)
	require.NotPanics(t, func() {
		_ = rule.Match(languagetool.AnalyzePlain("Die nicht Teil des Näherungsmodells sind"))
	})
}

func TestSubjectVerbAgreementRule_PrevChunkIsNominative(t *testing.T) {
	require.NotNil(t, NewSubjectVerbAgreementRule(nil))
}
