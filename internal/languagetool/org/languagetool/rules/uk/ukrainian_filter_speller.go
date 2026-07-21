package uk

import (
	"sync"

	atticmorfo "github.com/lucasew/lang/internal/attic/morfologik"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

var (
	ukFilterDictMu sync.RWMutex
	ukFilterDict   *atticmorfo.Dictionary
)

// WireUkrainianFilterSpeller opens CFSA2 dict for IsMisspelled / Suggest.
func WireUkrainianFilterSpeller(dictPath string) bool {
	if tools.JavaStringTrim(dictPath) == "" {
		return false
	}
	d, err := atticmorfo.OpenDictionary(dictPath)
	if err != nil || d == nil {
		return false
	}
	ukFilterDictMu.Lock()
	ukFilterDict = d
	ukFilterDictMu.Unlock()
	return true
}

// ClearUkrainianFilterSpeller clears wired dict (tests).
func ClearUkrainianFilterSpeller() {
	ukFilterDictMu.Lock()
	ukFilterDict = nil
	ukFilterDictMu.Unlock()
}

func getUKFilterDict() *atticmorfo.Dictionary {
	ukFilterDictMu.RLock()
	defer ukFilterDictMu.RUnlock()
	return ukFilterDict
}

// FilterDictIsMisspelledUK ports isMisspelled against wired dict; false when unwired.
func FilterDictIsMisspelledUK(word string) bool {
	d := getUKFilterDict()
	if d == nil || word == "" {
		return false
	}
	// Java Speller.isMisspelled (.info gates + Contains + convertCase)
	return d.IsMisspelled(word)
}

// FilterDictSuggestUK returns edit-distance suggestions from wired dict.
func FilterDictSuggestUK(word string) []string {
	d := getUKFilterDict()
	if d == nil || word == "" {
		return nil
	}
	return d.SuggestEdits(word, 8)
}

// FilterDictAvailableUK reports whether filter dict is wired.
func FilterDictAvailableUK() bool {
	return getUKFilterDict() != nil
}
