package pt

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMorfologikPortugueseSpellerRule(t *testing.T) {
	pt := NewMorfologikPortugalPortugueseSpellerRule()
	require.Equal(t, MorfologikPortuguesePTSpellerRuleID, pt.GetID())
	br := NewMorfologikBrazilianPortugueseSpellerRule()
	require.Equal(t, MorfologikPortugueseBRSpellerRuleID, br.GetID())
}
