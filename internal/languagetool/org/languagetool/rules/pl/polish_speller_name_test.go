package pl

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMorfologikPolishSpellerExists(t *testing.T) {
	r := NewMorfologikPolishSpellerRule()
	require.Equal(t, MorfologikPolishSpellerRuleID, r.GetID())
}
