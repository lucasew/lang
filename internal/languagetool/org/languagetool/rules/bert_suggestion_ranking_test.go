package rules

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBERTSuggestionRanking(t *testing.T) {
	b := NewBERTSuggestionRanking()
	b.Scorer = EditDistanceBERTScorer
	got := b.RankSuggestions("hello", []string{"helo", "hello", "hallo"}, nil, 0)
	require.Equal(t, "hello", got[0])

	// cache hit
	got2 := b.RankSuggestions("hello", []string{"helo", "hello", "hallo"}, nil, 0)
	require.Equal(t, got, got2)

	reps := []*SuggestedReplacement{
		NewSuggestedReplacement("helo"),
		NewSuggestedReplacement("hello"),
	}
	ranked := b.RankSuggestedReplacements("hello", reps, nil, 0)
	require.Equal(t, "hello", ranked[0].Replacement)
}
