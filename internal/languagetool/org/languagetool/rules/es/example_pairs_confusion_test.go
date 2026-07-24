package es

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestES_Confusion_ExamplePair(t *testing.T) {
	require.Equal(t, []string{"tuvo"}, NewSpanishConfusionProbabilityRule(nil).GetIncorrectExamples()[0].GetCorrections())
}
