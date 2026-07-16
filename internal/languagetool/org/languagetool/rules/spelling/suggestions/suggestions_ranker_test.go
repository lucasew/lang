package suggestions

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/stretchr/testify/require"
)

func TestThresholdSuggestionsRanker(t *testing.T) {
	r := NewThresholdSuggestionsRanker(0.9)
	require.False(t, r.IsMlAvailable())

	c1, c2 := float32(0.95), float32(0.8)
	require.True(t, r.ShouldAutoCorrect([]*rules.SuggestedReplacement{
		{Replacement: "a", Confidence: &c1},
		{Replacement: "b", Confidence: &c2},
	}))
	require.False(t, r.ShouldAutoCorrect([]*rules.SuggestedReplacement{
		{Replacement: "a", Confidence: &c2},
	}))
	require.False(t, r.ShouldAutoCorrect(nil))
}
