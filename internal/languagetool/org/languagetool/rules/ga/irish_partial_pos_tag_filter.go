package ga

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"

// IrishPartialPosTagFilter ports org.languagetool.rules.ga.IrishPartialPosTagFilter.
// Full tagger+disambiguator wiring is optional via Tag.
type IrishPartialPosTagFilter struct {
	*rules.PartialPosTagFilter
}

func NewIrishPartialPosTagFilter(tag func(string) []string) *IrishPartialPosTagFilter {
	return &IrishPartialPosTagFilter{PartialPosTagFilter: rules.NewPartialPosTagFilter(tag)}
}

// NoDisambiguationIrishPartialPosTagFilter ports
// org.languagetool.rules.ga.NoDisambiguationIrishPartialPosTagFilter.
type NoDisambiguationIrishPartialPosTagFilter struct {
	*rules.PartialPosTagFilter
}

func NewNoDisambiguationIrishPartialPosTagFilter(tag func(string) []string) *NoDisambiguationIrishPartialPosTagFilter {
	return &NoDisambiguationIrishPartialPosTagFilter{PartialPosTagFilter: rules.NewPartialPosTagFilter(tag)}
}
