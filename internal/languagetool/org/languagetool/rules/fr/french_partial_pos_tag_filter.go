package fr

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"

// FrenchPartialPosTagFilter wraps PartialPosTagFilter; full language tagger wiring is deferred.
type FrenchPartialPosTagFilter struct {
	*rules.PartialPosTagFilter
}

func NewFrenchPartialPosTagFilter(tag func(string) []string) *FrenchPartialPosTagFilter {
	return &FrenchPartialPosTagFilter{PartialPosTagFilter: rules.NewPartialPosTagFilter(tag)}
}
