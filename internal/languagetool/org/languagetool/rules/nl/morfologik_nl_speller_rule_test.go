package nl

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMorfologikDutchSpellerRule(t *testing.T) {
	require.Equal(t, MorfologikDutchSpellerRuleID, NewMorfologikDutchSpellerRule().GetID())
}
