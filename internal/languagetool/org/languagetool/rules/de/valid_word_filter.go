package de

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// ValidWordFilter ports org.languagetool.rules.de.ValidWordFilter.
// Suppresses a match when word1+word2 forms a known-good spelling.
// Java: GermanyGerman.getInstance().getDefaultSpellingRule().isMisspelled(...).
//
// IsMisspelled overrides the process-wide filter speller when set (tests).
// Without override and without WireGermanFilterSpeller → treat as misspelled
// (keep match; do not invent that the joined form is a valid word).
type ValidWordFilter struct {
	// IsMisspelled optional override; nil uses FilterDictIsMisspelled when available.
	IsMisspelled func(word string) bool
}

func NewValidWordFilter() *ValidWordFilter {
	return &ValidWordFilter{}
}

func (f *ValidWordFilter) isMisspelled(word string) bool {
	if f != nil && f.IsMisspelled != nil {
		return f.IsMisspelled(word)
	}
	if FilterDictAvailable() {
		return FilterDictIsMisspelled(word)
	}
	// No dict: always misspelled → keep match (fail-closed for suppress path).
	return true
}

// Accept returns true if the pattern match should be kept.
func (f *ValidWordFilter) Accept(word1, word2 string) bool {
	w1 := word1 + word2
	w2 := word1 + strings.ToLower(word2)
	// Java: if !isMisspelled(word1) || !isMisspelled(word2) return null (suppress)
	if !f.isMisspelled(w1) || !f.isMisspelled(w2) {
		return false
	}
	return true
}

// AcceptRuleMatch ports ValidWordFilter.acceptRuleMatch.
func (f *ValidWordFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if match == nil {
		return nil
	}
	if !f.Accept(arguments["word1"], arguments["word2"]) {
		return nil
	}
	return match
}
