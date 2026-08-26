package en

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// NoDisambiguationEnglishPartialPosTagFilter ports
// org.languagetool.rules.en.NoDisambiguationEnglishPartialPosTagFilter
// (PartialPosTagFilter + EnglishTagger only, no disambiguator).
//
// Java: Languages.getLanguageForShortCode("en").getTagger().tag(singletonList(token))
// — no getDisambiguator(). When tag is nil, uses process-wide filter tagger from
// WireEnglishFilterTagger. Without a wired tagger, Accept fails closed (no invent POS).
type NoDisambiguationEnglishPartialPosTagFilter struct {
	*rules.PartialPosTagFilter
}

// NewNoDisambiguationEnglishPartialPosTagFilter constructs the filter.
// tag maps a partial token to POS tags (injected for tests); nil → englishNoDisambigTagPOS.
func NewNoDisambiguationEnglishPartialPosTagFilter(tag func(string) []string) *NoDisambiguationEnglishPartialPosTagFilter {
	if tag == nil {
		tag = englishNoDisambigTagPOS
	}
	return &NoDisambiguationEnglishPartialPosTagFilter{
		PartialPosTagFilter: rules.NewPartialPosTagFilter(tag),
	}
}

// englishNoDisambigTagPOS ports NoDisambiguationEnglishPartialPosTagFilter.tag
// via the wired English filter tagger (Java: Languages.getLanguageForShortCode("en").getTagger()).
// Does not run the disambiguator (contrast EnglishPartialPosTagFilter).
func englishNoDisambigTagPOS(partial string) []string {
	tw := getFilterTagWord()
	if tw == nil {
		return nil
	}
	tags := tw(partial)
	out := make([]string, 0, len(tags))
	for _, t := range tags {
		if t.POS != "" {
			out = append(out, t.POS)
		}
	}
	return out
}
