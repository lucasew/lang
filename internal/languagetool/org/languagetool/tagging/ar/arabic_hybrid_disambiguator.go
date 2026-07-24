package ar

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// ArabicHybridDisambiguator ports org.languagetool.tagging.ar.ArabicHybridDisambiguator:
// MultiWordChunker.getInstance("/ar/multiwords.txt") defaults (F,F,F;
// NO setRemovePreviousTags; NO setIgnoreSpelling), then
// XmlRuleDisambiguator(new Arabic()) with useGlobalDisambiguation=false.
// Java order: disambiguator.disambiguate(chunker.disambiguate(input)) — multiwords then XML.
// Official ar/multiwords.txt is comment-only; Java still constructs MultiWordChunker —
// Chunker is eagerly wired (empty maps OK) when the official file is discoverable.
// Do not invent multiword entries.
type ArabicHybridDisambiguator struct {
	disambiguation.AbstractDisambiguator
	// Chunker is Java MultiWordChunker.getInstance("/ar/multiwords.txt") defaults.
	Chunker interface {
		Disambiguate(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
	}
	// Rules is Java XmlRuleDisambiguator (field name "disambiguator" in Java).
	Rules interface {
		Disambiguate(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
	}
}

// NewArabicHybridDisambiguator builds stages Java constructs as final fields.
// Chunker is wired from official /ar/multiwords.txt when discoverable
// (getInstance("/ar/multiwords.txt") defaults: F/F/F; no removePreviousTags; no ignoreSpelling).
// Official multiwords may be comment-only → non-nil empty MultiWordChunker (Java still constructs).
// Rules is eagerly wired from official ar/disambiguation.xml when present
// (XmlRuleDisambiguator(Arabic) — useGlobalDisambiguation=false).
func NewArabicHybridDisambiguator() *ArabicHybridDisambiguator {
	d := &ArabicHybridDisambiguator{}
	if mw := ArabicMultiWordChunker(); mw != nil {
		d.Chunker = mw
	}
	if xml := ArabicXmlRuleDisambiguator(); xml != nil {
		d.Rules = xml
	}
	return d
}

// NewArabicHybridDisambiguatorWithStages matches older call sites that pass stages.
func NewArabicHybridDisambiguatorWithStages(chunker, secondary disambiguation.Disambiguator) *ArabicHybridDisambiguator {
	return &ArabicHybridDisambiguator{Chunker: chunker, Rules: secondary}
}

// NewDefaultArabicHybridDisambiguator matches Java default construction (both stages when present).
func NewDefaultArabicHybridDisambiguator() *ArabicHybridDisambiguator {
	return NewArabicHybridDisambiguator()
}

func (d *ArabicHybridDisambiguator) Disambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if d == nil || input == nil {
		return input
	}
	out := input
	// Java: return disambiguator.disambiguate(chunker.disambiguate(input));
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
