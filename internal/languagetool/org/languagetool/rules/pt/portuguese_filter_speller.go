package pt

import (
	"sync"

	atticmorfo "github.com/lucasew/lang/internal/attic/morfologik"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
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

// TryWirePortugueseFilterSpeller discovers Java MorfologikPortugueseSpellerRule
// dictionaries (/pt/spelling/pt-PT-90.dict, pt-BR.dict, …) and wires the first openable.
func TryWirePortugueseFilterSpeller() bool {
	if FilterDictAvailable() {
		return true
	}
	// Order: pt-PT-90 (default Portuguese), then BR, then other variants if present.
	for _, rel := range []string{
		"pt/spelling/pt-PT-90.dict",
		"pt/spelling/pt-BR.dict",
		"pt/spelling/pt-PT-45.dict",
		"pt/spelling/pt-AO.dict",
	} {
		p := spelling.DiscoverSpellingResource(rel)
		if p != "" && WirePortugueseFilterSpeller(p) {
			return true
		}
	}
	return false
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
