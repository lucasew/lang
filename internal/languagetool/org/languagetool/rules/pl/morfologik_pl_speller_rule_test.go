package pl

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMorfologikPolishSpellerRule(t *testing.T) {
	require.Equal(t, MorfologikPolishSpellerRuleID, NewMorfologikPolishSpellerRule().GetID())
}
