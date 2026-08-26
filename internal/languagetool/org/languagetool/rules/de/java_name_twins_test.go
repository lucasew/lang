package de

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestJavaNameTwins(t *testing.T) {
	require.True(t, (GermanTools{}).IsVowel('ä'))
	require.Equal(t, "NOM", (GermanHelper{}).GetNounCase("SUB:NOM:SIN:MAS"))
	require.Contains(t, (PrepositionToCases{}).Cases("mit"), CaseDat)
	require.Greater(t, (CaseRuleAntiPatternsList{}).Count(), 0)
	require.NotEmpty(t, (AgreementRuleAntiPatterns1Data{}).Patterns())
}
