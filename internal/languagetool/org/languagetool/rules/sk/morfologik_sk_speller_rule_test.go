package sk

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMorfologikSlovakSpellerRule(t *testing.T) {
	require.Equal(t, MorfologikSlovakSpellerRuleID, NewMorfologikSlovakSpellerRule().GetID())
}
