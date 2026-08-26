package pl

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// PolishHybridDisambiguator ports org.languagetool.tagging.disambiguation.pl.PolishHybridDisambiguator:
//
//	chunker = MultiWordChunker.getInstance("/pl/multiwords.txt"); // defaults F,F,F
//	disambiguator = new XmlRuleDisambiguator(new Polish()); // useGlobalDisambiguation=false
//
// Java disambiguate order (nested call):
//
//	return chunker.disambiguate(disambiguator.disambiguate(input));
//
// i.e. XmlRuleDisambiguator FIRST, MultiWordChunker SECOND (inverted vs Romance hybrids).
// Both stages are eagerly wired from the same official resources Java loads.
type PolishHybridDisambiguator struct {
	disambiguation.AbstractDisambiguator
	// Chunker is Java MultiWordChunker.getInstance("/pl/multiwords.txt") defaults.
	Chunker interface {
		Disambiguate(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
	}
	// Rules is Java XmlRuleDisambiguator (field name "disambiguator" in Java).
	Rules interface {
		Disambiguate(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
	}
}

// NewPolishHybridDisambiguator ports Java field init:
//
//	chunker = MultiWordChunker.getInstance("/pl/multiwords.txt"); // F,F,F defaults
//	disambiguator = new XmlRuleDisambiguator(new Polish()); // useGlobalDisambiguation=false
//
// Stages are wired when the same official resources Java loads are discoverable.
func NewPolishHybridDisambiguator() *PolishHybridDisambiguator {
	d := &PolishHybridDisambiguator{}
	if mw := PolishMultiWordChunker(); mw != nil {
		d.Chunker = mw
	}
	if xml := PolishXmlRuleDisambiguator(); xml != nil {
		d.Rules = xml
	}
	return d
}

// NewPolishHybridDisambiguatorWithStages matches call sites that pass stages.
func NewPolishHybridDisambiguatorWithStages(chunker, secondary disambiguation.Disambiguator) *PolishHybridDisambiguator {
	return &PolishHybridDisambiguator{Chunker: chunker, Rules: secondary}
}

func (d *PolishHybridDisambiguator) Disambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if d == nil || input == nil {
		return input
	}
	out := input
	// Java PolishHybridDisambiguator:
	// return chunker.disambiguate(disambiguator.disambiguate(input));
	// i.e. XML rules first, then multiword chunker (inverted vs Romance hybrids).
	if d.Rules != nil {
		out = d.Rules.Disambiguate(out)
	}
	if d.Chunker != nil {
		out = d.Chunker.Disambiguate(out)
	}
	return out
}

var _ disambiguation.Disambiguator = (*PolishHybridDisambiguator)(nil)
