package crh

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMorfologikCrimeanTatarSpellerRule(t *testing.T) {
	r := NewMorfologikCrimeanTatarSpellerRule()
	require.Equal(t, MorfologikCrimeanTatarSpellerRuleID, r.GetID())
	require.Equal(t, MorfologikCrimeanTatarSpellerRuleDict, r.GetFileName())
}
