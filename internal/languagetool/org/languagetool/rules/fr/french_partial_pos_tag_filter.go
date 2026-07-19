package fr

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"

// FrenchPartialPosTagFilter ports org.languagetool.rules.fr.FrenchPartialPosTagFilter
// (PartialPosTagFilter that tags and disambiguates a single token in Java).
// Wire Tag to French tagger+disambiguator; nil Tag → fail-closed (drop matches).
type FrenchPartialPosTagFilter struct {
	*rules.PartialPosTagFilter
}

func NewFrenchPartialPosTagFilter(tag func(string) []string) *FrenchPartialPosTagFilter {
	return &FrenchPartialPosTagFilter{PartialPosTagFilter: rules.NewPartialPosTagFilter(tag)}
}
