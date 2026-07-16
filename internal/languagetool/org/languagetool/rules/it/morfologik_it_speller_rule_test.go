package it

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMorfologikItalianSpellerRule(t *testing.T) {
	r := NewMorfologikItalianSpellerRule()
	require.Equal(t, MorfologikItalianSpellerRuleID, r.GetID())
}
