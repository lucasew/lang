package morfologik

import (
	"sort"
	"strconv"
)

// WeightedSuggestion ports org.languagetool.rules.spelling.morfologik.WeightedSuggestion.
// Java fields are private final; exported Word/Weight are the Go struct layout with getters.
type WeightedSuggestion struct {
	Word   string
	Weight int
}

// NewWeightedSuggestion ports WeightedSuggestion(String word, int weight).
// Java: this.word = Objects.requireNonNull(word); empty strings are allowed (null only is rejected).
func NewWeightedSuggestion(word string, weight int) WeightedSuggestion {
	return WeightedSuggestion{Word: word, Weight: weight}
}

// GetWord ports getWord().
func (w WeightedSuggestion) GetWord() string { return w.Word }

// GetWeight ports getWeight().
func (w WeightedSuggestion) GetWeight() int { return w.Weight }

// CompareTo ports compareTo: Integer.compare(this.weight, o.getWeight()) — ascending by weight only.
func (w WeightedSuggestion) CompareTo(o WeightedSuggestion) int {
	if w.Weight < o.Weight {
		return -1
	}
	if w.Weight > o.Weight {
		return 1
	}
	return 0
}

// String ports toString: word + "/" + weight.
func (w WeightedSuggestion) String() string {
	return w.Word + "/" + strconv.Itoa(w.Weight)
}

// SortByWeight ports Collections.sort(List<WeightedSuggestion>):
// stable ascending sort by compareTo (weight only). Java Collections.sort is TimSort (stable);
// equal weights preserve relative order.
func SortByWeight(s []WeightedSuggestion) {
	sort.SliceStable(s, func(i, j int) bool {
		return s[i].CompareTo(s[j]) < 0
	})
}
