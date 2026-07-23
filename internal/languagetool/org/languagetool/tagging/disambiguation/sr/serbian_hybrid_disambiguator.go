package sr

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// SerbianHybridDisambiguator ports
// org.languagetool.tagging.disambiguation.sr.SerbianHybridDisambiguator:
// MultiWordChunker("/sr/multiwords.txt") defaults, then XmlRuleDisambiguator(Serbian) no global.
// Java order: disambiguator.disambiguate(chunker.disambiguate(input)) — multiwords then XML.
// Rules is eagerly wired from official sr/disambiguation.xml when present (Java final field).
type SerbianHybridDisambiguator struct {
	disambiguation.AbstractDisambiguator
	Chunker interface {
		Disambiguate(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
	}
	// Rules is Java XmlRuleDisambiguator (field name "disambiguator" in Java).
	Rules interface {
		Disambiguate(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
	}
}

// NewSerbianHybridDisambiguator ports Java field init: XmlRuleDisambiguator(new Serbian())
// (useGlobalDisambiguation=false). Chunker is left for injectors / multiword load helpers
// (same pattern as Swedish/Polish hybrid: multiwords loaded by callers).
func NewSerbianHybridDisambiguator() *SerbianHybridDisambiguator {
	d := &SerbianHybridDisambiguator{}
	if xml := SerbianXmlRuleDisambiguator(); xml != nil {
		d.Rules = xml
	}
	return d
}

// NewSerbianHybridDisambiguatorWithStages matches call sites that pass stages.
func NewSerbianHybridDisambiguatorWithStages(chunker, secondary disambiguation.Disambiguator) *SerbianHybridDisambiguator {
	return &SerbianHybridDisambiguator{Chunker: chunker, Rules: secondary}
}

func (d *SerbianHybridDisambiguator) Disambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if input == nil {
		return nil
	}
	out := input
	// multiwords first, then XML (Java: disambiguator.disambiguate(chunker.disambiguate(input)))
	if d.Chunker != nil {
		out = d.Chunker.Disambiguate(out)
	}
	if d.Rules != nil {
		out = d.Rules.Disambiguate(out)
	}
	return out
}

var _ disambiguation.Disambiguator = (*SerbianHybridDisambiguator)(nil)
