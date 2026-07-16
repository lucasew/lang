package fr

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMorfologikFrenchSpellerRule(t *testing.T) {
	r := NewMorfologikFrenchSpellerRule()
	require.Equal(t, MorfologikFrenchSpellerRuleID, r.GetID())
	require.Equal(t, FrenchSpellerDict, r.GetFileName())
}
