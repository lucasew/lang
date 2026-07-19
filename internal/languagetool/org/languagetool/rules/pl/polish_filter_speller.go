package pl

import (
	"strings"
	"sync"

	atticmorfo "github.com/lucasew/lang/internal/attic/morfologik"
)

// polish_filter_speller wires the official Polish Morfologik speller dict
// (/pl/hunspell/pl_PL.dict) into MorfologikPolishSpellerRule IsMisspelled.

var (
	filterDictMu sync.RWMutex
	filterDict   *atticmorfo.Dictionary
)

// WirePolishFilterSpeller opens a CFSA2/FSA speller dictionary.
// Returns false if path cannot be opened (speller stays fail-closed).
func WirePolishFilterSpeller(dictPath string) bool {
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

// ClearPolishFilterSpeller clears the process-wide filter dictionary (tests).
func ClearPolishFilterSpeller() {
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
// Without a dict, returns false (fail-closed: no invent misspell flags).
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
