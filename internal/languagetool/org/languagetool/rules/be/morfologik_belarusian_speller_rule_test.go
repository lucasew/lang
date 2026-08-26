package be

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMorfologikBelarusianSpellerRule(t *testing.T) {
	r := NewMorfologikBelarusianSpellerRule()
	require.Equal(t, MorfologikBelarusianSpellerRuleID, r.GetID())
	require.Equal(t, MorfologikBelarusianSpellerRuleDict, r.GetFileName())
}
