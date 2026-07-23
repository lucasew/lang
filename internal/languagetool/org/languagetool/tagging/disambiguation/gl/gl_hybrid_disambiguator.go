package gl

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// GalicianHybridDisambiguator ports org.languagetool.tagging.disambiguation.gl.GalicianHybridDisambiguator:
// MultiWordChunker.getInstance("/gl/multiwords.txt") defaults, then XmlRuleDisambiguator(Galician) no global.
// Java order: disambiguator.disambiguate(chunker.disambiguate(input)) — multiwords first, then XML.
// Both stages are eagerly wired from official resources when present (Java final fields).
type GalicianHybridDisambiguator struct {
	disambiguation.AbstractDisambiguator
	// Chunker is Java MultiWordChunker.getInstance("/gl/multiwords.txt") defaults.
	Chunker interface {
		Disambiguate(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
	}
	// Rules is Java XmlRuleDisambiguator (field name "disambiguator" in Java).
	Rules interface {
		Disambiguate(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
	}
}

// NewGalicianHybridDisambiguator ports Java field init:
//
//	chunker = MultiWordChunker.getInstance("/gl/multiwords.txt"); // F,F,F defaults
//	disambiguator = new XmlRuleDisambiguator(new Galician()); // useGlobalDisambiguation=false
//
// Stages are wired when the same official resources Java loads are discoverable.
func NewGalicianHybridDisambiguator() *GalicianHybridDisambiguator {
	d := &GalicianHybridDisambiguator{}
	if mw := GalicianMultiWordChunker(); mw != nil {
		d.Chunker = mw
	}
	if xml := GalicianXmlRuleDisambiguator(); xml != nil {
		d.Rules = xml
	}
	return d
}

// NewGalicianHybridDisambiguatorWithStages matches call sites that pass stages.
func NewGalicianHybridDisambiguatorWithStages(chunker, secondary disambiguation.Disambiguator) *GalicianHybridDisambiguator {
	return &GalicianHybridDisambiguator{Chunker: chunker, Rules: secondary}
}

func (d *GalicianHybridDisambiguator) Disambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if d == nil || input == nil {
		return input
	}
	out := input
	// Java GalicianHybridDisambiguator:
	// return disambiguator.disambiguate(chunker.disambiguate(input));
	// i.e. multiword chunker first, then XML rules (Romance order; inverted vs Polish/Swedish).
	if d.Chunker != nil {
		out = d.Chunker.Disambiguate(out)
	}
	if d.Rules != nil {
		out = d.Rules.Disambiguate(out)
	}
	return out
}

var _ disambiguation.Disambiguator = (*GalicianHybridDisambiguator)(nil)
