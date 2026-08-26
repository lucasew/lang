package de

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEstimateContextForSureMatch_JavaZeros(t *testing.T) {
	// Java WiederVsWiderRule / VerbAgreementRule return 0.
	require.Equal(t, 0, NewWiederVsWiderRule(nil).EstimateContextForSureMatch())
	require.Equal(t, 0, NewVerbAgreementRule(nil).EstimateContextForSureMatch())
}

func TestEstimateContextForSureMatch_AntiPatternMax(t *testing.T) {
	// Agreement / Case / SubjectVerb use max anti-pattern length (> 0 when table non-empty).
	require.Greater(t, NewAgreementRule(nil).EstimateContextForSureMatch(), 0)
	require.Greater(t, NewAgreementRule2(nil).EstimateContextForSureMatch(), 0)
	require.Greater(t, NewCaseRule(nil).EstimateContextForSureMatch(), 0)
	require.Greater(t, NewSubjectVerbAgreementRule(nil).EstimateContextForSureMatch(), 0)
}
