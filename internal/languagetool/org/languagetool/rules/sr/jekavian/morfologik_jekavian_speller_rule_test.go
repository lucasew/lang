package jekavian

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMorfologikJekavianSpellerRule(t *testing.T) {
	r := NewMorfologikJekavianSpellerRule()
	require.Equal(t, MorfologikJekavianSpellerRuleID, r.GetID())
	require.Equal(t, JekavianSpellerDict, r.GetFileName())
}
