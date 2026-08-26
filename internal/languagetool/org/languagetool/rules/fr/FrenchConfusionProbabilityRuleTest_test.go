package fr

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ngrams"
	"github.com/stretchr/testify/require"
)

func TestFrenchConfusionProbabilityRule_Constructor(t *testing.T) {
	r := NewFrenchConfusionProbabilityRule(ngrams.UniformLanguageModel(0.5, 1))
	require.NotNil(t, r)
	require.NotNil(t, r.ConfusionProbabilityRule)
}
