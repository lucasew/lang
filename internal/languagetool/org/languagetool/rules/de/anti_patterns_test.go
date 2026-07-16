package de

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAgreementAntiPatterns(t *testing.T) {
	require.NotEmpty(t, AgreementRuleAntiPatterns1)
	require.NotEmpty(t, AgreementRuleAntiPatterns2)
	require.NotEmpty(t, AgreementRuleAntiPatterns3)
	all := AllAgreementAntiPatterns()
	require.GreaterOrEqual(t, len(all), 3)
	require.Contains(t, MonthNamesRegex, "Januar")
	require.Greater(t, CaseRuleAntiPatternsCount(), 0)
	// first anti-pattern length
	require.Len(t, AgreementRuleAntiPatterns1[0], 6)
}

func TestGermanSpellerVariants(t *testing.T) {
	require.Equal(t, "AUSTRIAN_GERMAN_SPELLER_RULE", NewAustrianGermanSpellerRule(nil).GetID())
	require.Equal(t, "SWISS_GERMAN_SPELLER_RULE", NewSwissGermanSpellerRule(nil).GetID())
	r := NewMorfologikGermanyGermanSpellerRule(nil)
	require.Equal(t, "GERMAN_SPELLER_RULE", r.GetID())
	require.Equal(t, MorfologikGermanyGermanDict, r.GetMorfologikDictFilename())
	require.Equal(t, AustrianGermanSpellingDict, "de/hunspell/spelling-de-AT.txt")
}
