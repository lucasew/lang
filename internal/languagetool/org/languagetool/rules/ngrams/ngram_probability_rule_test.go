package ngrams

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestNgramProbabilityRule(t *testing.T) {
	lm := UniformLanguageModel(1e-20, 1)
	r := NewNgramProbabilityRule(lm)
	require.Equal(t, NgramRuleID, r.GetID())
	sent := languagetool.AnalyzePlain("hello world test")
	require.NotEmpty(t, r.Match(sent))

	r2 := NewNgramProbabilityRule(UniformLanguageModel(0.5, 1))
	require.Empty(t, r2.Match(sent))
}
