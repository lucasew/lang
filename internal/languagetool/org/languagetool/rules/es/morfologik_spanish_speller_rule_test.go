package es

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMorfologikSpanishSpellerRule(t *testing.T) {
	r := NewMorfologikSpanishSpellerRule()
	// Java MorfologikSpanishSpellerRule.getId / getFileName
	require.Equal(t, "MORFOLOGIK_RULE_ES", MorfologikSpanishSpellerRuleID)
	require.Equal(t, "/es/es-ES.dict", SpanishSpellerDict)
	require.Equal(t, MorfologikSpanishSpellerRuleID, r.GetID())
	require.Equal(t, SpanishSpellerDict, r.GetFileName())
}

func TestSpanishTokenizeNewWordsFalse(t *testing.T) {
	r := NewMorfologikSpanishSpellerRule()
	require.True(t, r.DisableTokenizeNewWords)
}
