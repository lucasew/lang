package ml

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMorfologikMalayalamSpellerRule(t *testing.T) {
	r := NewMorfologikMalayalamSpellerRule()
	require.Equal(t, MorfologikMalayalamSpellerRuleID, r.GetID())
	require.Equal(t, MorfologikMalayalamSpellerRuleDict, r.GetFileName())
}
