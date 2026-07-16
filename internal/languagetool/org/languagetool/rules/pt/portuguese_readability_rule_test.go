package pt

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPortugueseReadabilityRule_IDs(t *testing.T) {
	easy := NewPortugueseReadabilityRule(true, 2)
	require.Equal(t, "READABILITY_RULE_SIMPLE_PT", easy.GetID())
	require.Contains(t, easy.GetDescription(), "simples")
	hard := NewPortugueseReadabilityRule(false, 4)
	require.Equal(t, "READABILITY_RULE_DIFFICULT_PT", hard.GetID())
}
