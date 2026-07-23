package sv

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// SwedishHybridDisambiguator ports org.languagetool.tagging.disambiguation.sv.SwedishHybridDisambiguator:
// MultiWordChunker.getInstance("/sv/multiwords.txt") defaults, then XmlRuleDisambiguator(Swedish) no global.
// Java order: chunker.disambiguate(disambiguator.disambiguate(input)) — XML first, then multiwords.
// Both stages are eagerly wired from official resources when present (Java final fields).
type SwedishHybridDisambiguator struct {
	disambiguation.AbstractDisambiguator
	// Chunker is Java MultiWordChunker.getInstance("/sv/multiwords.txt") defaults.
	Chunker interface {
		Disambiguate(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
	}
	// Rules is Java XmlRuleDisambiguator (field name "disambiguator" in Java).
	Rules interface {
		Disambiguate(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
	}
}

// NewSwedishHybridDisambiguator ports Java field init:
//
//	chunker = MultiWordChunker.getInstance("/sv/multiwords.txt"); // F,F,F defaults
//	disambiguator = new XmlRuleDisambiguator(new Swedish()); // useGlobalDisambiguation=false
//
// Stages are wired when the same official resources Java loads are discoverable.
func NewSwedishHybridDisambiguator() *SwedishHybridDisambiguator {
	d := &SwedishHybridDisambiguator{}
	if mw := SwedishMultiWordChunker(); mw != nil {
		d.Chunker = mw
	}
	if xml := SwedishXmlRuleDisambiguator(); xml != nil {
		d.Rules = xml
	}
	return d
}

// NewSwedishHybridDisambiguatorWithStages matches call sites that pass stages.
func NewSwedishHybridDisambiguatorWithStages(chunker, secondary disambiguation.Disambiguator) *SwedishHybridDisambiguator {
	return &SwedishHybridDisambiguator{Chunker: chunker, Rules: secondary}
}

func (d *SwedishHybridDisambiguator) Disambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if d == nil || input == nil {
		return input
	}
	out := input
	// Java SwedishHybridDisambiguator:
	// return chunker.disambiguate(disambiguator.disambiguate(input));
	// i.e. XML rules first, then multiword chunker (same inverted order as Polish).
	if d.Rules != nil {
		out = d.Rules.Disambiguate(out)
	}
	if d.Chunker != nil {
		out = d.Chunker.Disambiguate(out)
	}
	return out
}

var _ disambiguation.Disambiguator = (*SwedishHybridDisambiguator)(nil)
