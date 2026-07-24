package tl

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMorfologikTagalogSpellerRule(t *testing.T) {
	r := NewMorfologikTagalogSpellerRule()
	// Java MorfologikTagalogSpellerRule.getId / getFileName
	require.Equal(t, "MORFOLOGIK_RULE_TL", MorfologikTagalogSpellerRuleID)
	require.Equal(t, "/tl/hunspell/tl_PH.dict", TagalogSpellerDict)
	require.Equal(t, MorfologikTagalogSpellerRuleID, r.GetID())
	require.Equal(t, TagalogSpellerDict, r.GetFileName())
}
