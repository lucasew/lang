package pt

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"

// NoDisambiguationPortuguesePartialPosTagFilter wraps PartialPosTagFilter; full language tagger wiring is deferred.
type NoDisambiguationPortuguesePartialPosTagFilter struct {
	*rules.PartialPosTagFilter
}

func NewNoDisambiguationPortuguesePartialPosTagFilter(tag func(string) []string) *NoDisambiguationPortuguesePartialPosTagFilter {
	return &NoDisambiguationPortuguesePartialPosTagFilter{PartialPosTagFilter: rules.NewPartialPosTagFilter(tag)}
}
