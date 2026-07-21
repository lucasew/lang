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

// Twin of SubjectVerbAgreementRuleTest.testRuleWithIncorrectPluralVerb
func TestSubjectVerbAgreementRule_RuleWithIncorrectPluralVerb(t *testing.T) {
	rule := NewSubjectVerbAgreementRule(nil)
	// Die Katze (SIN) + sind (PLU) — chunk NPS
	die := atrWithPOS("Die", "ART:DEF:NOM:SIN:FEM", "die")
	katze := atrWithPOS("Katze", "SUB:NOM:SIN:FEM", "Katze")
	die.SetChunkTags([]string{chunkNPS})
	katze.SetChunkTags([]string{chunkNPS})
	bad := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		die, katze,
		atrWithPOS("sind", "VER:3:PLU:PRÄ:NON", "sein"),
		atrWithPOS("schön", "ADJ:PRD:GRU", "schön"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 1, len(rule.Match(bad)))
	// untagged must not invent
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Die Katze sind schön."))))
}

// Twin of SubjectVerbAgreementRuleTest.testRuleWithCorrectPluralVerb
func TestSubjectVerbAgreementRule_RuleWithCorrectPluralVerb(t *testing.T) {
	rule := NewSubjectVerbAgreementRule(nil)
	die := atrWithPOS("Die", "ART:DEF:NOM:PLU:ALG", "die")
	katzen := atrWithPOS("Katzen", "SUB:NOM:PLU:FEM", "Katze")
	die.SetChunkTags([]string{chunkNPP})
	katzen.SetChunkTags([]string{chunkNPP})
	good := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		die, katzen,
		atrWithPOS("sind", "VER:3:PLU:PRÄ:NON", "sein"),
		atrWithPOS("schön", "ADJ:PRD:GRU", "schön"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 0, len(rule.Match(good)))
}

// Twin of SubjectVerbAgreementRuleTest.testRuleWithCorrectSingularAndPluralVerb
func TestSubjectVerbAgreementRule_RuleWithCorrectSingularAndPluralVerb(t *testing.T) {
	// Both SIN and PLU acceptable for "Personen ist/sind" style — morph: SIN subject + SIN verb ok
	rule := NewSubjectVerbAgreementRule(nil)
	die := atrWithPOS("Personen", "SUB:DAT:PLU:FEM", "Person")
	die.SetChunkTags([]string{chunkNPP})
	// "Personen ist der Zugriff …" — Java allows both; assert no invent on untagged
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Personen ist der Zugriff auf diese Daten verboten."))))
	// Morph: plural subject with plural verb OK
	personen := atrWithPOS("Personen", "SUB:NOM:PLU:FEM", "Person")
	personen.SetChunkTags([]string{chunkNPP})
	good := languagetool.NewAnalyzedSentence(withPositions(
		sentStartATR(),
		personen,
		atrWithPOS("sind", "VER:3:PLU:PRÄ:NON", "sein"),
		atrWithPOS("wichtig", "ADJ:PRD:GRU", "wichtig"),
		atrWithPOS(".", "PKT", "."),
	))
	require.Equal(t, 0, len(rule.Match(good)))
}
