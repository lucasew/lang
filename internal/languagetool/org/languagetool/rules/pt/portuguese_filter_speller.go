package pt

import (
	"strings"
	"sync"

	atticmorfo "github.com/lucasew/lang/internal/attic/morfologik"
)

// portuguese_filter_speller wires the official Portuguese Morfologik speller dict into
// grammar RuleFilters (PortugueseSuppressMisspelledSuggestionsFilter, etc.).
// Java MorfologikPortugueseSpellerRule uses /pt/spelling/{pt-BR|pt-PT-90|…}.dict.

var (
	filterDictMu sync.RWMutex
	filterDict   *atticmorfo.Dictionary
)

// WirePortugueseFilterSpeller opens a CFSA2/FSA speller dictionary for filter hooks.
// Returns false if path cannot be opened (filters stay fail-closed / null-speller semantics).
func WirePortugueseFilterSpeller(dictPath string) bool {
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

// ClearPortugueseFilterSpeller clears the process-wide filter dictionary (tests).
func ClearPortugueseFilterSpeller() {
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
