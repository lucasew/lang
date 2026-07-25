package de

import (
	"sync"
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
		// InsertCommaFilter.getTag → GermanTagger.INSTANCE.tag (full tagger).
		SetDefaultInsertCommaGermanTagger(tagger)
		// UppercaseNounReadingFilter → GermanTagger.INSTANCE.tag (full tagger, not POS invent).
		SetDefaultUppercaseNounGermanTagger(tagger)
	})
}
