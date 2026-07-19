package ca

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCA_ExamplePairs(t *testing.T) {
	// Multi-marker wrong: first fixed marker is pesebre
	require.Equal(t, []string{"pesebre"}, NewWordCoherencyRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Equal(t, []string{"Ryanair"}, NewCompoundRule(nil).GetIncorrectExamples()[0].GetCorrections())
	require.Contains(t, NewCatalanWordRepeatBeginningRule(nil).GetIncorrectExamples()[0].GetExample(), "Però")
}
