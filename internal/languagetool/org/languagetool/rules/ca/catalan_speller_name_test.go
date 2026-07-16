package ca

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMorfologikCatalanSpellerExists(t *testing.T) {
	r := NewMorfologikCatalanSpellerRule()
	require.Equal(t, MorfologikCatalanSpellerRuleID, r.GetID())
}
