package de

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDE_Confusion_ExamplePair(t *testing.T) {
	require.Equal(t, []string{"mit"}, NewGermanConfusionProbabilityRule(nil).GetIncorrectExamples()[0].GetCorrections())
}
