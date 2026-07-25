package de

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Twin-oriented checks for GermanReadabilityRule (no soft ≥0 smoke).
// Core algorithm / match behavior covered in german_readability_rule_test.go.

func TestGermanReadabilityRule_SyllablesAndFRE(t *testing.T) {
	r := NewGermanReadabilityRule(nil, false)
	require.Equal(t, 1, r.SimpleSyllablesCount("a"))
	require.Equal(t, 1, r.SimpleSyllablesCount("Haus"))
	require.InDelta(t, 180-10-58.5*1.5, r.GetFleschReadingEase(10, 1.5), 0.01)
	// readability levels from base (same thresholds as English ReadabilityRule)
	require.Equal(t, 0, r.GetReadabilityLevel(10))
	require.Equal(t, 6, r.GetReadabilityLevel(95))
}

func TestGermanReadabilityRule_IDs(t *testing.T) {
	require.Equal(t, "READABILITY_RULE_DIFFICULT_DE", NewGermanReadabilityRule(nil, false).GetID())
	require.Equal(t, "READABILITY_RULE_SIMPLE_DE", NewGermanReadabilityRule(nil, true).GetID())
}
