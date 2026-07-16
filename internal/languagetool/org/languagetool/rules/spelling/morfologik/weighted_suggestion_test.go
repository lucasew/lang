package morfologik

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWeightedSuggestion(t *testing.T) {
	a := NewWeightedSuggestion("hello", 2)
	b := NewWeightedSuggestion("hi", 1)
	require.True(t, b.Less(a))
	require.Equal(t, "hi/1", b.String())
	list := []WeightedSuggestion{a, b}
	SortByWeight(list)
	require.Equal(t, "hi", list[0].Word)
}
