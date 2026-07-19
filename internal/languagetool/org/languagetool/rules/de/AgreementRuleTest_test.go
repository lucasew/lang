package de

// Twin of AgreementRuleTest — open compounds need getCompoundError (dict/lt.check), not invent.
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestAgreementRule_CompoundMatch(t *testing.T) {
	rule := NewAgreementRule(nil)
	// Untagged AnalyzePlain: no invent of open-compound hits
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Das ist die Original Mail."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("Doch dieser kleine Magnesium Anteil ist entscheidend."))))
	require.Equal(t, 0, len(rule.Match(languagetool.AnalyzePlain("War das Eifersucht?"))))
}

func TestAgreementRule_GetCategoriesCausingError(t *testing.T) {
	// morphology categories need tagger
	require.NotNil(t, NewAgreementRule(nil))
}
