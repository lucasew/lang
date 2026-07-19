package en

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"

// EnglishPartialPosTagFilter ports org.languagetool.rules.en.EnglishPartialPosTagFilter
// (PartialPosTagFilter that tags and disambiguates a single token in Java).
// Without a Tag hook (tagger+disambiguator), Accept fails closed.
// Do not invent disambiguation when only a bare tagger is available.
type EnglishPartialPosTagFilter struct {
	*rules.PartialPosTagFilter
}

func NewEnglishPartialPosTagFilter(tag func(string) []string) *EnglishPartialPosTagFilter {
	return &EnglishPartialPosTagFilter{PartialPosTagFilter: rules.NewPartialPosTagFilter(tag)}
}

// NoDisambiguationEnglishPartialPosTagFilter ports
// org.languagetool.rules.en.NoDisambiguationEnglishPartialPosTagFilter
// (PartialPosTagFilter + EnglishTagger only, no disambiguator).
// When tag is nil, uses the process-wide filter tagger from WireEnglishFilterTagger.
// Without a wired tagger, Accept fails closed (do not invent POS).
type NoDisambiguationEnglishPartialPosTagFilter struct {
	*rules.PartialPosTagFilter
}

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
