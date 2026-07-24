package nl

import (
	"sync"

	atticmorfo "github.com/lucasew/lang/internal/attic/morfologik"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// dutch_filter_tagger wires official dutch.dict POS lookup for CompoundAcceptor
// (Java DutchTagger.getPostags → getWordTagger().tag only — no compound re-entry).

const dutchPOSDictClasspath = "/nl/dutch.dict"

var (
	filterTagMu   sync.RWMutex
	filterTagDict *atticmorfo.Dictionary
)

// WireDutchFilterTagger opens CFSA2 dutch.dict for getPostags-style probes.
// Returns false if path cannot be opened (CompoundAcceptor POS stays fail-closed).
func WireDutchFilterTagger(dictPath string) bool {
	if tools.JavaStringTrim(dictPath) == "" {
		return false
	}
	d, err := atticmorfo.OpenDictionary(dictPath)
	if err != nil || d == nil {
		return false
	}
	filterTagMu.Lock()
	filterTagDict = d
	filterTagMu.Unlock()
	// Java CompoundAcceptor holds DutchTagger.INSTANCE; attach when dict opens.
	BindDefaultCompoundAcceptorFilters()
	return true
}

// TryWireDutchFilterTagger discovers Java /nl/dutch.dict and wires it when present.
func TryWireDutchFilterTagger() bool {
	p := morfologik.DiscoverLanguageDict(dutchPOSDictClasspath)
	if p == "" {
		return false
	}
	return WireDutchFilterTagger(p)
}

// ClearDutchFilterTagger clears the process-wide filter POS dict (tests).
func ClearDutchFilterTagger() {
	filterTagMu.Lock()
	filterTagDict = nil
	filterTagMu.Unlock()
}

func getFilterTagDict() *atticmorfo.Dictionary {
	filterTagMu.RLock()
	defer filterTagMu.RUnlock()
	return filterTagDict
}

// FilterTaggerAvailable reports whether dutch.dict is wired.
func FilterTaggerAvailable() bool {
	return getFilterTagDict() != nil
}

// FilterGetPostags ports DutchTagger.getPostags: raw word-tagger tags only
// (does not run compound acceptance — avoids tagger↔CompoundAcceptor loop).
// Without a dict, returns nil (fail-closed).
func FilterGetPostags(word string) []string {
	d := getFilterTagDict()
	if d == nil || word == "" {
		return nil
	}
	forms, err := d.Lookup(word)
	if err != nil || len(forms) == 0 {
		return nil
	}
	out := make([]string, 0, len(forms))
	for _, f := range forms {
		if f.Tag != "" {
			out = append(out, f.Tag)
		}
	}
	return out
}
