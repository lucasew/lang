package en

import (
	"sync"

	atticmorfo "github.com/lucasew/lang/internal/attic/morfologik"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// english_filter_speller wires Morfologik dictionaries into grammar RuleFilters.
//
// Two dict roles (mirrors Java):
//   - filterDict: locale hunspell dict (e.g. en_US.dict) for isMisspelled /
//     NumberInWord / SuppressMisspelled — same as MorfologikAmericanSpellerRule.
//   - findSuggestionsDict: /en/english.dict for FindSuggestionsFilter
//     MorfologikSpeller.findSimilarWords (Java DICT_FILENAME).

var (
	filterDictMu        sync.RWMutex
	filterDict          *atticmorfo.Dictionary
	findSuggestionsDict *atticmorfo.Dictionary
)

// WireEnglishFilterSpeller opens a CFSA2/FSA speller dictionary for filter isMisspelled hooks.
// Returns false if path cannot be opened (filters stay fail-closed / no misspell probe).
func WireEnglishFilterSpeller(dictPath string) bool {
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

// WireEnglishFindSuggestionsSpeller opens /en/english.dict for FindSuggestionsFilter
// (Java: new MorfologikSpeller("/en/english.dict") when resourceExists).
// Called from WireEnglishFilterTagger (same official resource path).
func WireEnglishFindSuggestionsSpeller(dictPath string) bool {
	if tools.JavaStringTrim(dictPath) == "" {
		return false
	}
	d, err := atticmorfo.OpenDictionary(dictPath)
	if err != nil || d == nil {
		return false
	}
	filterDictMu.Lock()
	findSuggestionsDict = d
	filterDictMu.Unlock()
	return true
}

// ClearEnglishFilterSpeller clears the process-wide misspell filter dictionary (tests).
func ClearEnglishFilterSpeller() {
	filterDictMu.Lock()
	filterDict = nil
	filterDictMu.Unlock()
}

// ClearEnglishFindSuggestionsSpeller clears the FindSuggestions MorfologikSpeller dict (tests).
func ClearEnglishFindSuggestionsSpeller() {
	filterDictMu.Lock()
	findSuggestionsDict = nil
	filterDictMu.Unlock()
}

func getFilterDict() *atticmorfo.Dictionary {
	filterDictMu.RLock()
	defer filterDictMu.RUnlock()
	return filterDict
}

func getFindSuggestionsDict() *atticmorfo.Dictionary {
	filterDictMu.RLock()
	defer filterDictMu.RUnlock()
	return findSuggestionsDict
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

// FilterDictSuggest ports FindSuggestionsFilter.getSpellingSuggestions:
// MorfologikSpeller.findSimilarWords(token) on /en/english.dict when wired.
// Falls back to the misspell filter dict if only that is present (host wiring).
// Without any dict, returns nil (fail-closed — do not invent similar words).
func FilterDictSuggest(word string) []string {
	d := getFindSuggestionsDict()
	if d == nil {
		d = getFilterDict()
	}
	if d == nil || word == "" {
		return nil
	}
	// Java MorfologikSpeller default maxEditDistance=1; findSimilarWords (evenIfInDictionary).
	return d.FindSimilarWords(word, 1)
}

// FilterDictAvailable reports whether a misspell filter speller dict is wired.
func FilterDictAvailable() bool {
	return getFilterDict() != nil
}

// FilterFindSuggestionsAvailable reports whether FindSuggestions can call findSimilarWords
// (english.dict and/or filter dict wired). Java: speller non-null when resourceExists.
func FilterFindSuggestionsAvailable() bool {
	return getFindSuggestionsDict() != nil || getFilterDict() != nil
}
