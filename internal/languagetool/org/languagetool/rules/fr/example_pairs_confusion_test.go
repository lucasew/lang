package fr

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFR_Confusion_ExamplePair(t *testing.T) {
	require.Equal(t, []string{"prix"}, NewFrenchConfusionProbabilityRule(nil).GetIncorrectExamples()[0].GetCorrections())
}
