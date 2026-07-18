package en

import (
	"strings"
	"sync"

	atticmorfo "github.com/lucasew/lang/internal/attic/morfologik"
)

// english_filter_speller wires the official en_US Morfologik dict into grammar
// RuleFilters (NumberInWord, FindSuggestions, SuppressMisspelled) and multitoken
// isMisspelled — same resource Java MorfologikAmericanSpellerRule uses.

var (
	filterDictMu sync.RWMutex
	filterDict   *atticmorfo.Dictionary
)

// WireEnglishFilterSpeller opens a CFSA2/FSA speller dictionary for filter hooks.
// Returns false if path cannot be opened (filters stay fail-closed / no misspell probe).
func WireEnglishFilterSpeller(dictPath string) bool {
	if strings.TrimSpace(dictPath) == "" {
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

// ClearEnglishFilterSpeller clears the process-wide filter dictionary (tests).
func ClearEnglishFilterSpeller() {
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
	if d.Contains(word) {
		return false
	}
	low := strings.ToLower(word)
	if low != word && d.Contains(low) {
		return false
	}
	return true
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
