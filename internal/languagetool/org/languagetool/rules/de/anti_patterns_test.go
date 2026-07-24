package de

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAgreementAntiPatterns(t *testing.T) {
	// Java AgreementRuleAntiPatterns{1,2,3}.ANTI_PATTERNS lengths (indent-4 asList count).
	require.Len(t, AgreementRuleAntiPatterns1, 165)
	require.Len(t, AgreementRuleAntiPatterns2, 135)
	require.Len(t, AgreementRuleAntiPatterns3, 127)
	all := AllAgreementAntiPatterns()
	require.Len(t, all, 165+135+127)
	require.Contains(t, MonthNamesRegex, "Januar")
	require.Greater(t, CaseRuleAntiPatternsCount(), 0)
	// first anti-pattern length
	require.Len(t, AgreementRuleAntiPatterns1[0], 6)
	// GermanWordRepeatRule.ANTI_PATTERNS (indent-4 Arrays.asList count)
	require.Len(t, GermanWordRepeatAntiPatterns, 59)
}

func TestGermanSpellerVariants(t *testing.T) {
	require.Equal(t, "AUSTRIAN_GERMAN_SPELLER_RULE", NewAustrianGermanSpellerRule(nil).GetID())
	require.Equal(t, "SWISS_GERMAN_SPELLER_RULE", NewSwissGermanSpellerRule(nil).GetID())
	r := NewMorfologikGermanyGermanSpellerRule(nil)
	require.Equal(t, "MORFOLOGIK_RULE_DE_DE", r.GetID())
	require.Equal(t, MorfologikGermanyGermanDict, r.GetMorfologikDictFilename())
	require.Equal(t, AustrianGermanSpellingDict, "de/hunspell/spelling-de-AT.txt")
}
