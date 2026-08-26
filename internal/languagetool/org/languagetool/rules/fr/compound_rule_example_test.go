package fr

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Java CompoundRule: Haut Rhin → Haut-Rhin
func TestCompoundRule_ExamplePair(t *testing.T) {
	r := NewCompoundRule(nil)
	require.Equal(t, "FR_COMPOUNDS", r.GetID())
	require.Equal(t, []string{"Haut-Rhin"}, r.GetIncorrectExamples()[0].GetCorrections())
}
