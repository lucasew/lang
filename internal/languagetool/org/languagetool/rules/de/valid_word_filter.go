package de

import "strings"

// ValidWordFilter ports org.languagetool.rules.de.ValidWordFilter.
// Suppresses a match when word1+word2 forms a known-good spelling.
type ValidWordFilter struct {
	// IsMisspelled reports unknown words; nil defaults to always misspelled
	// (never suppress without a dictionary).
	IsMisspelled func(word string) bool
}

func NewValidWordFilter() *ValidWordFilter {
	return &ValidWordFilter{}
}

// Accept returns true if the pattern match should be kept.
func (f *ValidWordFilter) Accept(word1, word2 string) bool {
	miss := f.IsMisspelled
	if miss == nil {
		miss = func(string) bool { return true }
	}
	w1 := word1 + word2
	w2 := word1 + strings.ToLower(word2)
	// If either spelling is valid, suppress (return false)
	if !miss(w1) || !miss(w2) {
		return false
	}
	return true
}
