package de

// Twin of VerbAgreementRuleTest (surface pronoun+sein forms).
import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestVerbAgreementRule_WrongVerb(t *testing.T) {
	rule := NewVerbAgreementRule(nil)
	matchN := func(s string) int {
		return len(rule.Match(languagetool.AnalyzePlain(s)))
	}
	require.Equal(t, 0, matchN("Du bist in dem Moment angekommen, als ich gegangen bin."))
	require.Equal(t, 0, matchN("Ich bin müde."))
	require.Equal(t, 1, matchN("Ich sind müde."))
	require.Equal(t, 0, matchN("Die Jagd nach bin Laden."))
}

func TestVerbAgreementRule_SuggestionSorting(t *testing.T) {
	require.NotNil(t, NewVerbAgreementRule(nil))
}

func TestVerbAgreementRule_Positions(t *testing.T) {
	rule := NewVerbAgreementRule(nil)
	ms := rule.Match(languagetool.AnalyzePlain("Du ist hier."))
	require.Equal(t, 1, len(ms))
}
