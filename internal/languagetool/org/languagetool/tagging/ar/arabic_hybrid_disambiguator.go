package ar

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// ArabicHybridDisambiguator ports org.languagetool.tagging.ar.ArabicHybridDisambiguator:
// MultiWordChunker.getInstance("/ar/multiwords.txt") defaults, then XmlRuleDisambiguator(Arabic) no global.
// Java order: disambiguator.disambiguate(chunker.disambiguate(input)) — multiwords then XML.
// Official ar/multiwords.txt is comment-only; Chunker stays nil unless injected (no invent entries).
// Rules is eagerly wired from official ar/disambiguation.xml when present (Java final field).
type ArabicHybridDisambiguator struct {
	disambiguation.AbstractDisambiguator
	Chunker interface {
		Disambiguate(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
	}
	// Rules is Java XmlRuleDisambiguator (field name "disambiguator" in Java).
	Rules interface {
		Disambiguate(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
	}
}

// NewArabicHybridDisambiguator ports Java field init: XmlRuleDisambiguator(new Arabic())
// (useGlobalDisambiguation=false). MultiWordChunker is left for injectors / commandline
// (official multiwords empty — no invent).
func NewArabicHybridDisambiguator() *ArabicHybridDisambiguator {
	d := &ArabicHybridDisambiguator{}
	if xml := ArabicXmlRuleDisambiguator(); xml != nil {
		d.Rules = xml
	}
	return d
}

// NewArabicHybridDisambiguatorWithStages matches older call sites that pass stages.
func NewArabicHybridDisambiguatorWithStages(chunker, secondary disambiguation.Disambiguator) *ArabicHybridDisambiguator {
	return &ArabicHybridDisambiguator{Chunker: chunker, Rules: secondary}
}

// NewDefaultArabicHybridDisambiguator matches Java default construction (XML rules stage).
func NewDefaultArabicHybridDisambiguator() *ArabicHybridDisambiguator {
	return NewArabicHybridDisambiguator()
}

func (d *ArabicHybridDisambiguator) Disambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if d == nil || input == nil {
		return input
	}
	out := input
	// multiwords first, then XML
	if d.Chunker != nil {
		out = d.Chunker.Disambiguate(out)
	}
	if d.Rules != nil {
		out = d.Rules.Disambiguate(out)
	}
	return out
}

func (d *ArabicHybridDisambiguator) PreDisambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	return input
}

var _ disambiguation.Disambiguator = (*ArabicHybridDisambiguator)(nil)
