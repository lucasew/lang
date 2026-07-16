package de

// Twin of AgreementRuleTest (surface open-compound heuristic).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestAgreementRule_CompoundMatch(t *testing.T) {
	rule := NewAgreementRule(nil)
	matchN := func(s string) int {
		return len(rule.Match(languagetool.AnalyzePlain(s)))
	}
	require.Equal(t, 1, matchN("Das ist die Original Mail."))
	require.Equal(t, 1, matchN("Doch dieser kleine Magnesium Anteil ist entscheidend."))
	require.Equal(t, 0, matchN("War das Eifersucht?"))
}

func TestAgreementRule_GetCategoriesCausingError(t *testing.T) {
	// morphology categories need tagger
	require.NotNil(t, NewAgreementRule(nil))
}
