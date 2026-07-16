package ekavian

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMorfologikEkavianSpellerRule(t *testing.T) {
	r := NewMorfologikEkavianSpellerRule()
	require.Equal(t, MorfologikEkavianSpellerRuleID, r.GetID())
	require.Equal(t, EkavianSpellerDict, r.GetFileName())
}
