package ru

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// NoDisambiguationRussianPartialPosTagFilter ports
// org.languagetool.rules.ru.NoDisambiguationRussianPartialPosTagFilter
// (PartialPosTagFilter + RussianTagger only, no disambiguator).
//
// Java: Languages.getLanguageForShortCode("ru").getTagger().tag(singletonList(token))
// — no getDisambiguator(). When tag is nil, uses process-wide filter tagger from
// SetDefaultRussianPartialPosTagger / WireRussianFilterTaggerFromTagWord.
// Without a wired tagger, Accept fails closed (no invent POS).
//
// Contrast: RussianPartialPosTagFilter runs hybrid disambiguator on the one-token sentence.
type NoDisambiguationRussianPartialPosTagFilter struct {
	*rules.PartialPosTagFilter
}

// NewNoDisambiguationRussianPartialPosTagFilter constructs the filter.
// tag maps a partial token to POS tags (injected for tests); nil → russianNoDisambigTagPOS.
func NewNoDisambiguationRussianPartialPosTagFilter(tag func(string) []string) *NoDisambiguationRussianPartialPosTagFilter {
	if tag == nil {
		tag = russianNoDisambigTagPOS
	}
	return &NoDisambiguationRussianPartialPosTagFilter{
		PartialPosTagFilter: rules.NewPartialPosTagFilter(tag),
	}
}

// russianNoDisambigTagPOS ports NoDisambiguationRussianPartialPosTagFilter.tag
// via the wired Russian filter tagger (Java: Languages.getLanguageForShortCode("ru").getTagger()).
// Does not run the disambiguator (contrast RussianPartialPosTagFilter).
func russianNoDisambigTagPOS(partial string) []string {
	tag := getDefaultRussianPartialPosTagger()
	if tag == nil {
		return nil
	}
	return tag(partial)
}
