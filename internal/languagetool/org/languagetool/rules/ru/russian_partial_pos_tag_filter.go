package ru

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"

// RussianPartialPosTagFilter wraps PartialPosTagFilter; full language tagger wiring is deferred.
type RussianPartialPosTagFilter struct {
	*rules.PartialPosTagFilter
}

func NewRussianPartialPosTagFilter(tag func(string) []string) *RussianPartialPosTagFilter {
	return &RussianPartialPosTagFilter{PartialPosTagFilter: rules.NewPartialPosTagFilter(tag)}
}
