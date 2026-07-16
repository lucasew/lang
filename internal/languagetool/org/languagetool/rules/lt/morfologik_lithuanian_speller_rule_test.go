package lt

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMorfologikLithuanianSpellerRule(t *testing.T) {
	r := NewMorfologikLithuanianSpellerRule()
	require.Equal(t, MorfologikLithuanianSpellerRuleID, r.GetID())
	require.Equal(t, MorfologikLithuanianSpellerRuleDict, r.GetFileName())
}
