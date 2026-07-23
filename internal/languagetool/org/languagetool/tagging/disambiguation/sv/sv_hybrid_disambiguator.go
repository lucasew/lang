package sv

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// SwedishHybridDisambiguator ports org.languagetool.tagging.disambiguation.sv.SwedishHybridDisambiguator:
// MultiWordChunker.getInstance("/sv/multiwords.txt") defaults, then XmlRuleDisambiguator(Swedish) no global.
// Java order: chunker.disambiguate(disambiguator.disambiguate(input)) — XML first, then multiwords.
// Rules is eagerly wired from official sv/disambiguation.xml when present (Java final field).
type SwedishHybridDisambiguator struct {
	disambiguation.AbstractDisambiguator
	Chunker interface {
		Disambiguate(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
	}
	// Rules is Java XmlRuleDisambiguator (field name "disambiguator" in Java).
	Rules interface {
		Disambiguate(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
	}
}

// NewSwedishHybridDisambiguator ports Java field init: XmlRuleDisambiguator(new Swedish())
// (useGlobalDisambiguation=false). Chunker is left for injectors / multiword load helpers
// (same pattern as Polish hybrid: multiwords loaded by callers).
func NewSwedishHybridDisambiguator() *SwedishHybridDisambiguator {
	d := &SwedishHybridDisambiguator{}
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
	if input == nil {
		return nil
	}
	out := input
	if d.Rules != nil {
		out = d.Rules.Disambiguate(out)
	}
	if d.Chunker != nil {
		out = d.Chunker.Disambiguate(out)
	}
	return out
}

var _ disambiguation.Disambiguator = (*SwedishHybridDisambiguator)(nil)
