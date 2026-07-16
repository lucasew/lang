package uk

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMorfologikUkrainianSpellerRule(t *testing.T) {
	require.Equal(t, MorfologikUkrainianSpellerRuleID, NewMorfologikUkrainianSpellerRule().GetID())
}
