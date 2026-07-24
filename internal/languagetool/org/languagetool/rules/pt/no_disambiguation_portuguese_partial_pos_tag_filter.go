package pt

import (
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
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

// WirePortugueseFilterTaggerFromTagWord installs POS list hook from lt.TagWord
// (Java NoDisambiguationPortuguesePartialPosTagFilter uses Portuguese.getTagger()).
func WirePortugueseFilterTaggerFromTagWord(tw func(token string) []languagetool.TokenTag) {
	if tw == nil {
		SetDefaultPortuguesePartialPosTagger(nil)
		return
	}
	SetDefaultPortuguesePartialPosTagger(func(token string) []string {
		tags := tw(token)
		if len(tags) == 0 {
			return nil
		}
		out := make([]string, 0, len(tags))
		for _, t := range tags {
			if t.POS != "" {
				out = append(out, t.POS)
			}
		}
		return out
	})
}
