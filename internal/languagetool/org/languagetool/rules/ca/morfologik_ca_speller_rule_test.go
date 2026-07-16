package ca

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMorfologikCatalanSpellerRule(t *testing.T) {
	require.Equal(t, MorfologikCatalanSpellerRuleID, NewMorfologikCatalanSpellerRule().GetID())
}
