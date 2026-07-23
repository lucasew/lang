package pl

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// PolishHybridDisambiguator ports org.languagetool.tagging.disambiguation.pl.PolishHybridDisambiguator:
// MultiWordChunker.getInstance("/pl/multiwords.txt") defaults, then XmlRuleDisambiguator(Polish) no global.
// Java order: chunker.disambiguate(disambiguator.disambiguate(input)) — XML first, then multiwords.
// Rules is eagerly wired from official pl/disambiguation.xml when present (Java final field).
type PolishHybridDisambiguator struct {
	disambiguation.AbstractDisambiguator
	Chunker interface {
		Disambiguate(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
	}
	// Rules is Java XmlRuleDisambiguator (field name "disambiguator" in Java).
	Rules interface {
		Disambiguate(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
	}
}

// NewPolishHybridDisambiguator ports Java field init: XmlRuleDisambiguator(new Polish())
// (useGlobalDisambiguation=false). Chunker is left for injectors / multiword load helpers
// (same pattern as Russian hybrid: multiwords loaded by callers).
func NewPolishHybridDisambiguator() *PolishHybridDisambiguator {
	d := &PolishHybridDisambiguator{}
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
	// i.e. XML rules first, then multiword chunker.
	if d.Rules != nil {
		out = d.Rules.Disambiguate(out)
	}
	if d.Chunker != nil {
		out = d.Chunker.Disambiguate(out)
	}
	return out
}

var _ disambiguation.Disambiguator = (*PolishHybridDisambiguator)(nil)
