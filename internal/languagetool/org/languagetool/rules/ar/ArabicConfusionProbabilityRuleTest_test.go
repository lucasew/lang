package ar

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ngrams"
	"github.com/stretchr/testify/require"
)

func TestArabicConfusionProbabilityRule_Constructor(t *testing.T) {
	r := NewArabicConfusionProbabilityRule(ngrams.UniformLanguageModel(0.5, 1))
	require.NotNil(t, r)
}
