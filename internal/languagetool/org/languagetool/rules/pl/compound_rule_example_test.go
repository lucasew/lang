package pl

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Java CompoundRule: Rabce Zdroju → Rabce-Zdroju
func TestCompoundRule_ExamplePair(t *testing.T) {
	require.Equal(t, []string{"Rabce-Zdroju"}, NewCompoundRule(nil).GetIncorrectExamples()[0].GetCorrections())
}
