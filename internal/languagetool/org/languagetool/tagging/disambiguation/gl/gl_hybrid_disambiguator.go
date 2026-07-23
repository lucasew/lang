package gl

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// GalicianHybridDisambiguator ports org.languagetool.tagging.disambiguation.gl.GalicianHybridDisambiguator:
// MultiWordChunker.getInstance("/gl/multiwords.txt") defaults, then XmlRuleDisambiguator(Galician) no global.
// Java order: disambiguator.disambiguate(chunker.disambiguate(input)) — multiwords first, then XML.
// Rules is eagerly wired from official gl/disambiguation.xml when present (Java final field).
type GalicianHybridDisambiguator struct {
	disambiguation.AbstractDisambiguator
	Chunker interface {
		Disambiguate(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
	}
	// Rules is Java XmlRuleDisambiguator (field name "disambiguator" in Java).
	Rules interface {
		Disambiguate(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
	}
}

// NewGalicianHybridDisambiguator ports Java field init: XmlRuleDisambiguator(new Galician())
// (useGlobalDisambiguation=false). Chunker is left for injectors / multiword load helpers
// (same pattern as Russian hybrid: multiwords loaded by callers).
func NewGalicianHybridDisambiguator() *GalicianHybridDisambiguator {
	d := &GalicianHybridDisambiguator{}
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
	// multiwords first, then XML (Java GalicianHybridDisambiguator)
	if d.Chunker != nil {
		out = d.Chunker.Disambiguate(out)
	}
	if d.Rules != nil {
		out = d.Rules.Disambiguate(out)
	}
	return out
}

var _ disambiguation.Disambiguator = (*GalicianHybridDisambiguator)(nil)
