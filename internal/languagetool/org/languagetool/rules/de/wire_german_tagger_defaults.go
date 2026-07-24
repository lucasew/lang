package de

import (
	"sync"

	detag "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/de"
)

// WireGermanTaggerDefaults attaches process-wide GermanTagger.INSTANCE equivalents
// for grammar filters that call tag() on surface forms (InsertCommaFilter,
// UppercaseNounReadingFilter). Without resources, filter POS branches stay fail-closed.
var (
	taggerDefaultsOnce sync.Once
)

func WireGermanTaggerDefaults() {
	taggerDefaultsOnce.Do(func() {
		// Reuse process-wide DiscoveredGermanTagger (same openDiscovered stack).
		tagger := DiscoveredGermanTagger()
		if tagger == nil {
			return
		}
		tagPOS := germanTaggerPOSFunc(tagger)
		SetDefaultInsertCommaTagger(tagPOS)
		SetDefaultUppercaseNounTagger(tagPOS)
	})
}

func germanTaggerPOSFunc(tagger *detag.GermanTagger) func(string) []string {
	return func(word string) []string {
		if tagger == nil || word == "" {
			return nil
		}
		rd := tagger.Lookup(word)
		if rd == nil {
			return nil
		}
		var out []string
		for _, r := range rd.GetReadings() {
			if r == nil {
				continue
			}
			if p := r.GetPOSTag(); p != nil && *p != "" {
				out = append(out, *p)
			}
		}
		return out
	}
}
