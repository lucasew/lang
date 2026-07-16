package ru

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"

// NoDisambiguationRussianPartialPosTagFilter wraps PartialPosTagFilter; full language tagger wiring is deferred.
type NoDisambiguationRussianPartialPosTagFilter struct {
	*rules.PartialPosTagFilter
}

func NewNoDisambiguationRussianPartialPosTagFilter(tag func(string) []string) *NoDisambiguationRussianPartialPosTagFilter {
	return &NoDisambiguationRussianPartialPosTagFilter{PartialPosTagFilter: rules.NewPartialPosTagFilter(tag)}
}
