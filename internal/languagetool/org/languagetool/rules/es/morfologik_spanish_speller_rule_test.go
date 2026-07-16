package es

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMorfologikSpanishSpellerRule(t *testing.T) {
	r := NewMorfologikSpanishSpellerRule()
	require.Equal(t, MorfologikSpanishSpellerRuleID, r.GetID())
}
