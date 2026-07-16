package de

import "strings"

// RemoveUnknownCompoundsFilter ports RemoveUnknownCompoundsFilter.
// Suppresses match when part1+part2 is misspelled as a compound.
type RemoveUnknownCompoundsFilter struct {
	// IsMisspelled nil defaults to false (keep match without dictionary).
	IsMisspelled func(word string) bool
}

func NewRemoveUnknownCompoundsFilter() *RemoveUnknownCompoundsFilter {
	return &RemoveUnknownCompoundsFilter{}
}

// Accept returns true if the match should be kept.
func (f *RemoveUnknownCompoundsFilter) Accept(part1, part2 string) bool {
	miss := f.IsMisspelled
	if miss == nil {
		miss = func(string) bool { return false }
	}
	compound := part1 + strings.ToLower(part2)
	if miss(compound) {
		return false
	}
	return true
}
