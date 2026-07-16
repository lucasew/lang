package sl

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMorfologikSlovenianSpellerRule(t *testing.T) {
	r := NewMorfologikSlovenianSpellerRule()
	require.Equal(t, MorfologikSlovenianSpellerRuleID, r.GetID())
}
