package es

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ngrams"
	"github.com/stretchr/testify/require"
)

func TestSpanishConfusionProbabilityRule_Constructor(t *testing.T) {
	r := NewSpanishConfusionProbabilityRule(ngrams.UniformLanguageModel(0.5, 1))
	require.NotNil(t, r)
	require.NotNil(t, r.ConfusionProbabilityRule)
	require.Equal(t, ngrams.ConfusionRuleID, r.GetID())
}
