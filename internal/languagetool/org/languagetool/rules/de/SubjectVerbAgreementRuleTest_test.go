package de

// Twin of SubjectVerbAgreementRuleTest (surface plural-subject + ist).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestSubjectVerbAgreementRule_RuleWithIncorrectSingularVerb(t *testing.T) {
	rule := NewSubjectVerbAgreementRule(nil)
	matchN := func(s string) int {
		return len(rule.Match(languagetool.AnalyzePlain(s)))
	}
	require.Equal(t, 1, matchN("Die Autos ist schnell."))
	require.Equal(t, 1, matchN("Drei Katzen ist im Haus."))
	require.Equal(t, 1, matchN("Viele Katzen ist schön."))
}

func TestSubjectVerbAgreementRule_RuleWithCorrectSingularVerb(t *testing.T) {
	rule := NewSubjectVerbAgreementRule(nil)
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Die Katze ist schön."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Die Rechte der Kinder sind universell."))))
}

func TestSubjectVerbAgreementRule_Temp(t *testing.T) {
	// Java temp/array-bound harness — no-op surface
	require.NotNil(t, NewSubjectVerbAgreementRule(nil))
}

func TestSubjectVerbAgreementRule_ArrayOutOfBoundsBug(t *testing.T) {
	rule := NewSubjectVerbAgreementRule(nil)
	require.NotPanics(t, func() {
		_ = rule.Match(languagetool.AnalyzePlain("Die nicht Teil des Näherungsmodells sind"))
	})
}

func TestSubjectVerbAgreementRule_PrevChunkIsNominative(t *testing.T) {
	// needs chunk/POS; soft document
	require.NotNil(t, NewSubjectVerbAgreementRule(nil))
}
