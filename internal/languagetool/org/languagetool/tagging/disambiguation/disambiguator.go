package disambiguation

import "github.com/lucasew/lang/internal/languagetool/org/languagetool"

// Disambiguator ports org.languagetool.tagging.disambiguation.Disambiguator.
type Disambiguator interface {
	// PreDisambiguate runs before XML disambiguation rules.
	PreDisambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
	// Disambiguate filters incorrect POS tags.
	Disambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
}

// AbstractDisambiguator ports AbstractDisambiguator with identity pre-disambiguation.
type AbstractDisambiguator struct{}

func (AbstractDisambiguator) PreDisambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	return input
}
