package de

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMorfologikGermanyGermanSpellerRule(t *testing.T) {
	r := NewMorfologikGermanyGermanSpellerRule(nil)
	require.NotNil(t, r)
	// Java MorfologikGermanyGermanSpellerRule.getId / getFileName
	require.Equal(t, "MORFOLOGIK_RULE_DE_DE", MorfologikGermanyGermanSpellerRuleID)
	require.Equal(t, "/de/hunspell/de_DE.dict", MorfologikGermanyGermanDict)
	require.Equal(t, MorfologikGermanyGermanSpellerRuleID, r.GetID())
	require.Equal(t, MorfologikGermanyGermanDict, r.GetFileName())
	require.Equal(t, MorfologikGermanyGermanDict, r.GetMorfologikDictFilename())
}
