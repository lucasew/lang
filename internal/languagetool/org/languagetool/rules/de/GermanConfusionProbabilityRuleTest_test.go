package de

// Twin of GermanConfusionProbabilityRuleTest.
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGermanConfusionProbabilityRule_Constructor(t *testing.T) {
	r := NewGermanConfusionProbabilityRule(nil)
	require.NotNil(t, r)
	require.Equal(t, "DE_CONFUSION_RULE", r.GetID())
}
