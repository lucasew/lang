package ru

import (
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// RussianPartialPosTagFilter ports org.languagetool.rules.ru.RussianPartialPosTagFilter
// (PartialPosTagFilter that tags and disambiguates a single token in Java).
// Without Tag (tagger+disambiguator), Accept fails closed — do not invent disambiguation.
type RussianPartialPosTagFilter struct {
	*rules.PartialPosTagFilter
}

func NewRussianPartialPosTagFilter(tag func(string) []string) *RussianPartialPosTagFilter {
	return &RussianPartialPosTagFilter{PartialPosTagFilter: rules.NewPartialPosTagFilter(tag)}
}

// NoDisambiguationRussianPartialPosTagFilter ports
// org.languagetool.rules.ru.NoDisambiguationRussianPartialPosTagFilter
// (PartialPosTagFilter + Russian tagger only, no disambiguator).
// When tag is nil, uses the process-wide filter tagger from SetDefaultRussianPartialPosTagger.
// Without a wired tagger, Accept fails closed (do not invent POS).
type NoDisambiguationRussianPartialPosTagFilter struct {
	*rules.PartialPosTagFilter
}

func NewNoDisambiguationRussianPartialPosTagFilter(tag func(string) []string) *NoDisambiguationRussianPartialPosTagFilter {
	if tag == nil {
		tag = russianNoDisambigTagPOS
	}
	return &NoDisambiguationRussianPartialPosTagFilter{
		PartialPosTagFilter: rules.NewPartialPosTagFilter(tag),
	}
}

var (
	ruPartialTagMu sync.RWMutex
	ruPartialTag   func(string) []string
)

// SetDefaultRussianPartialPosTagger wires Russian tagger POS for NoDisambiguation filters.
// Java: Languages.getLanguageForShortCode("ru").getTagger().
func SetDefaultRussianPartialPosTagger(tag func(string) []string) {
	ruPartialTagMu.Lock()
	defer ruPartialTagMu.Unlock()
	ruPartialTag = tag
}

// ClearDefaultRussianPartialPosTagger clears the process-wide tagger (tests).
func ClearDefaultRussianPartialPosTagger() {
	SetDefaultRussianPartialPosTagger(nil)
}

func russianNoDisambigTagPOS(partial string) []string {
	ruPartialTagMu.RLock()
	tag := ruPartialTag
	ruPartialTagMu.RUnlock()
	if tag == nil {
		return nil
	}
	return tag(partial)
}
