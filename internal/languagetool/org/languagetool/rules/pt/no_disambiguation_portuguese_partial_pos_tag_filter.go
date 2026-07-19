package pt

import (
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// NoDisambiguationPortuguesePartialPosTagFilter ports
// org.languagetool.rules.pt.NoDisambiguationPortuguesePartialPosTagFilter
// (PartialPosTagFilter + Portuguese tagger only, no disambiguator).
// When tag is nil, uses the process-wide filter tagger from SetDefaultPortuguesePartialPosTagger.
// Without a wired tagger, Accept fails closed (do not invent POS).
type NoDisambiguationPortuguesePartialPosTagFilter struct {
	*rules.PartialPosTagFilter
}

func NewNoDisambiguationPortuguesePartialPosTagFilter(tag func(string) []string) *NoDisambiguationPortuguesePartialPosTagFilter {
	if tag == nil {
		tag = portugueseNoDisambigTagPOS
	}
	return &NoDisambiguationPortuguesePartialPosTagFilter{
		PartialPosTagFilter: rules.NewPartialPosTagFilter(tag),
	}
}

var (
	ptPartialTagMu sync.RWMutex
	ptPartialTag   func(string) []string
)

// SetDefaultPortuguesePartialPosTagger wires Portuguese tagger POS for this filter.
// Java: Portuguese.getInstance().getTagger().
func SetDefaultPortuguesePartialPosTagger(tag func(string) []string) {
	ptPartialTagMu.Lock()
	defer ptPartialTagMu.Unlock()
	ptPartialTag = tag
}

// ClearDefaultPortuguesePartialPosTagger clears the process-wide tagger (tests).
func ClearDefaultPortuguesePartialPosTagger() {
	SetDefaultPortuguesePartialPosTagger(nil)
}

func portugueseNoDisambigTagPOS(partial string) []string {
	ptPartialTagMu.RLock()
	tag := ptPartialTag
	ptPartialTagMu.RUnlock()
	if tag == nil {
		return nil
	}
	return tag(partial)
}
