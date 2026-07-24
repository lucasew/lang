package en

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/ngrams"
	"github.com/stretchr/testify/require"
)

func TestEnglishConfusionProbabilityRule_Constructor(t *testing.T) {
	r := NewEnglishConfusionProbabilityRule(ngrams.UniformLanguageModel(0.5, 1))
	require.NotNil(t, r)
	require.Equal(t, EnglishConfusionRuleID, r.GetID())
}
