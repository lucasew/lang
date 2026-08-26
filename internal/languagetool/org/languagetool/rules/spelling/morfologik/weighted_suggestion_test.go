package morfologik

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Behavior matrix for WeightedSuggestion (no Java @Test class):
// constructor/getters, compareTo ascending weight, toString "word/weight",
// Collections.sort via SortByWeight (stable on equal weights).

func TestWeightedSuggestion_Getters(t *testing.T) {
	w := NewWeightedSuggestion("hello", 42)
	require.Equal(t, "hello", w.GetWord())
	require.Equal(t, 42, w.GetWeight())
	require.Equal(t, "hello", w.Word)
	require.Equal(t, 42, w.Weight)
}

func TestWeightedSuggestion_EmptyWordAllowed(t *testing.T) {
	// Java Objects.requireNonNull rejects null only; empty string is valid.
	w := NewWeightedSuggestion("", 0)
	require.Equal(t, "", w.GetWord())
	require.Equal(t, 0, w.GetWeight())
	require.Equal(t, "/0", w.String())
}

func TestWeightedSuggestion_CompareTo_AscendingWeight(t *testing.T) {
	// Integer.compare(this.weight, o.getWeight())
	a := NewWeightedSuggestion("a", 1)
	b := NewWeightedSuggestion("b", 2)
	c := NewWeightedSuggestion("c", 1)
	require.Equal(t, -1, a.CompareTo(b))
	require.Equal(t, 1, b.CompareTo(a))
	require.Equal(t, 0, a.CompareTo(c))
	// weight only — word ignored for ordering
	require.Equal(t, 0, NewWeightedSuggestion("zzz", 5).CompareTo(NewWeightedSuggestion("aaa", 5)))
}

func TestWeightedSuggestion_String(t *testing.T) {
	require.Equal(t, "hi/1", NewWeightedSuggestion("hi", 1).String())
	require.Equal(t, "hello/2", NewWeightedSuggestion("hello", 2).String())
	require.Equal(t, "x/0", NewWeightedSuggestion("x", 0).String())
	require.Equal(t, "neg/-3", NewWeightedSuggestion("neg", -3).String())
	require.Equal(t, "big/1000", NewWeightedSuggestion("big", 1000).String())
}

func TestWeightedSuggestion_SortByWeight_Ascending(t *testing.T) {
	// Collections.sort uses compareTo → lower weight first.
	list := []WeightedSuggestion{
		NewWeightedSuggestion("hello", 2),
		NewWeightedSuggestion("hi", 1),
		NewWeightedSuggestion("hey", 10),
		NewWeightedSuggestion("a", 0),
	}
	SortByWeight(list)
	require.Equal(t, []string{"a", "hi", "hello", "hey"}, wordsOf(list))
	require.Equal(t, []int{0, 1, 2, 10}, weightsOf(list))
}

func TestWeightedSuggestion_SortByWeight_StableOnEqualWeights(t *testing.T) {
	// Java Collections.sort is stable TimSort; equal weights keep relative order.
	list := []WeightedSuggestion{
		NewWeightedSuggestion("first", 5),
		NewWeightedSuggestion("second", 1),
		NewWeightedSuggestion("third", 5),
		NewWeightedSuggestion("fourth", 1),
		NewWeightedSuggestion("fifth", 5),
	}
	SortByWeight(list)
	require.Equal(t, []string{"second", "fourth", "first", "third", "fifth"}, wordsOf(list))
	require.Equal(t, []int{1, 1, 5, 5, 5}, weightsOf(list))
}

func TestWeightedSuggestion_SortByWeight_AlreadySorted(t *testing.T) {
	list := []WeightedSuggestion{
		NewWeightedSuggestion("a", 1),
		NewWeightedSuggestion("b", 2),
		NewWeightedSuggestion("c", 3),
	}
	SortByWeight(list)
	require.Equal(t, []string{"a", "b", "c"}, wordsOf(list))
}

func TestWeightedSuggestion_SortByWeight_EmptyAndNil(t *testing.T) {
	SortByWeight(nil)
	SortByWeight([]WeightedSuggestion{})
	one := []WeightedSuggestion{NewWeightedSuggestion("only", 7)}
	SortByWeight(one)
	require.Equal(t, "only/7", one[0].String())
}

func wordsOf(list []WeightedSuggestion) []string {
	out := make([]string, len(list))
	for i, w := range list {
		out[i] = w.GetWord()
	}
	return out
}

func weightsOf(list []WeightedSuggestion) []int {
	out := make([]int, len(list))
	for i, w := range list {
		out[i] = w.GetWeight()
	}
	return out
}
