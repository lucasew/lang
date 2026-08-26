package de

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RemoveUnknownCompoundsFilter ports org.languagetool.rules.de.RemoveUnknownCompoundsFilter.
// Suppresses match when part1+part2 is misspelled as a compound.
// Java: GermanyGerman.getInstance().getDefaultSpellingRule().isMisspelled(compound).
//
// IsMisspelled overrides the process-wide filter speller when set (tests).
// Without an override and without WireGermanFilterSpeller → fail-closed drop
// (do not invent that a compound is spelled correctly).
type RemoveUnknownCompoundsFilter struct {
	// IsMisspelled optional override; nil uses FilterDictIsMisspelled when available.
	IsMisspelled func(word string) bool
}

func NewRemoveUnknownCompoundsFilter() *RemoveUnknownCompoundsFilter {
	return &RemoveUnknownCompoundsFilter{}
}

// isMisspelled resolves the speller like Java default spelling rule.
// ok=false means no speller available (fail-closed).
func (f *RemoveUnknownCompoundsFilter) isMisspelled(word string) (misspelled bool, ok bool) {
	if f != nil && f.IsMisspelled != nil {
		return f.IsMisspelled(word), true
	}
	if FilterDictAvailable() {
		return FilterDictIsMisspelled(word), true
	}
	return false, false
}

// Accept returns true if the match should be kept.
// Java: if isMisspelled(compound) return null; else return match.
func (f *RemoveUnknownCompoundsFilter) Accept(part1, part2 string) bool {
	if f == nil {
		return false
	}
	compound := part1 + strings.ToLower(part2)
	miss, ok := f.isMisspelled(compound)
	if !ok {
		// Without speller cannot verify compound is known (fail-closed: drop).
		return false
	}
	if miss {
		return false
	}
	return true
}

// AcceptRuleMatch ports RemoveUnknownCompoundsFilter.acceptRuleMatch.
func (f *RemoveUnknownCompoundsFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if match == nil {
		return nil
	}
	if !f.Accept(arguments["part1"], arguments["part2"]) {
		return nil
	}
	return match
}
