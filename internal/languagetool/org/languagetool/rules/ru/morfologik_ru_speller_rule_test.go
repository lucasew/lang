package ru

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMorfologikRussianSpellerRules(t *testing.T) {
	require.Equal(t, MorfologikRussianSpellerRuleID, NewMorfologikRussianSpellerRule().GetID())
	require.Equal(t, MorfologikRussianYOSpellerRuleID, NewMorfologikRussianYOSpellerRule().GetID())
}
