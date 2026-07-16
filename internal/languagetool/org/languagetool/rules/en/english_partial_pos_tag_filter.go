package en

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"

// EnglishPartialPosTagFilter wraps PartialPosTagFilter with a pluggable tagger.
// Full English tagger + disambiguator integration is deferred.
type EnglishPartialPosTagFilter struct {
	*rules.PartialPosTagFilter
}

func NewEnglishPartialPosTagFilter(tag func(string) []string) *EnglishPartialPosTagFilter {
	return &EnglishPartialPosTagFilter{PartialPosTagFilter: rules.NewPartialPosTagFilter(tag)}
}

// NoDisambiguationEnglishPartialPosTagFilter is the same surface type without disambiguator.
type NoDisambiguationEnglishPartialPosTagFilter struct {
	*rules.PartialPosTagFilter
}

func NewNoDisambiguationEnglishPartialPosTagFilter(tag func(string) []string) *NoDisambiguationEnglishPartialPosTagFilter {
	return &NoDisambiguationEnglishPartialPosTagFilter{PartialPosTagFilter: rules.NewPartialPosTagFilter(tag)}
}
