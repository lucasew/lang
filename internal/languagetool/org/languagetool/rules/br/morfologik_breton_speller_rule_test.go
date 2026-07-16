package br

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMorfologikBretonSpellerRule(t *testing.T) {
	r := NewMorfologikBretonSpellerRule()
	require.Equal(t, MorfologikBretonSpellerRuleID, r.GetID())
	require.Equal(t, MorfologikBretonSpellerRuleDict, r.GetFileName())
}
