package fr

import (
	"sync"

	atticmorfo "github.com/lucasew/lang/internal/attic/morfologik"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// french_filter_speller wires the official French Morfologik speller dict into
// grammar RuleFilters (FrenchNumberInWordFilter, etc.) — same resource Java
// MorfologikFrenchSpellerRule loads (/fr/french.dict).

var (
	filterDictMu sync.RWMutex
	filterDict   *atticmorfo.Dictionary
)

// WireFrenchFilterSpeller opens a CFSA2/FSA speller dictionary for filter hooks.
// Returns false if path cannot be opened (filters stay fail-closed).
func WireFrenchFilterSpeller(dictPath string) bool {
	if tools.JavaStringTrim(dictPath) == "" {
		return false
	}
	d, err := atticmorfo.OpenDictionary(dictPath)
	if err != nil || d == nil {
		return false
	}
	filterDictMu.Lock()
	filterDict = d
	filterDictMu.Unlock()
	return true
}

// ClearFrenchFilterSpeller clears the process-wide filter dictionary (tests).
func ClearFrenchFilterSpeller() {
	filterDictMu.Lock()
	filterDict = nil
	filterDictMu.Unlock()
}

func getFilterDict() *atticmorfo.Dictionary {
	filterDictMu.RLock()
	defer filterDictMu.RUnlock()
	return filterDict
}

// FilterDictIsMisspelled ports spelling-rule isMisspelled against the wired dict.
// Without a dict, returns false (Java: null SpellingCheckRule → isMisspelled false).
func FilterDictIsMisspelled(word string) bool {
	d := getFilterDict()
	if d == nil || word == "" {
		return false
	}
	// Java Speller.isMisspelled (.info gates + Contains + convertCase)
	return d.IsMisspelled(word)
}

// FilterDictSuggest returns edit-distance suggestions from the wired dict.
func FilterDictSuggest(word string) []string {
	d := getFilterDict()
	if d == nil || word == "" {
		return nil
	}
	return d.SuggestEdits(word, 8)
}

// FilterDictAvailable reports whether a filter speller dict is wired.
func FilterDictAvailable() bool {
	return getFilterDict() != nil
}
