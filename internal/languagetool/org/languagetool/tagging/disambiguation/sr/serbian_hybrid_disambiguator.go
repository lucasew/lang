package sr

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// SerbianHybridDisambiguator ports
// org.languagetool.tagging.disambiguation.sr.SerbianHybridDisambiguator:
// MultiWordChunker("/sr/multiwords.txt") with getInstance defaults (F,F,F;
// NO setRemovePreviousTags; NO setIgnoreSpelling), then
// XmlRuleDisambiguator(new Serbian()) with useGlobalDisambiguation=false.
// Java order: disambiguator.disambiguate(chunker.disambiguate(input)) — multiwords then XML.
// Official sr/multiwords.txt is empty (0 lines); Java still constructs MultiWordChunker —
// Chunker is eagerly wired (empty maps OK) when the official file is discoverable.
// Do not invent multiword entries.
type SerbianHybridDisambiguator struct {
	disambiguation.AbstractDisambiguator
	// Chunker is Java MultiWordChunker("/sr/multiwords.txt") / getInstance defaults.
	Chunker interface {
		Disambiguate(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
	}
	// Rules is Java XmlRuleDisambiguator (field name "disambiguator" in Java).
	Rules interface {
		Disambiguate(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
	}
}

// NewSerbianHybridDisambiguator builds stages Java constructs as final fields.
// Chunker is wired from official /sr/multiwords.txt when discoverable
// (getInstance defaults: F/F/F; no removePreviousTags; no ignoreSpelling).
// Official multiwords may be empty → non-nil empty MultiWordChunker (Java still constructs).
// Rules is eagerly wired from official sr/disambiguation.xml when present
// (XmlRuleDisambiguator(Serbian) — useGlobalDisambiguation=false).
func NewSerbianHybridDisambiguator() *SerbianHybridDisambiguator {
	d := &SerbianHybridDisambiguator{}
	if mw := SerbianMultiWordChunker(); mw != nil {
		d.Chunker = mw
	}
	if xml := SerbianXmlRuleDisambiguator(); xml != nil {
		d.Rules = xml
	}
	return d
}

// NewSerbianHybridDisambiguatorWithStages matches call sites that pass stages.
func NewSerbianHybridDisambiguatorWithStages(chunker, secondary disambiguation.Disambiguator) *SerbianHybridDisambiguator {
	return &SerbianHybridDisambiguator{Chunker: chunker, Rules: secondary}
}

// NewDefaultSerbianHybridDisambiguator matches Java default construction (both stages when present).
func NewDefaultSerbianHybridDisambiguator() *SerbianHybridDisambiguator {
	return NewSerbianHybridDisambiguator()
}

func (d *SerbianHybridDisambiguator) Disambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
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

func (d *SerbianHybridDisambiguator) PreDisambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	return input
}

var _ disambiguation.Disambiguator = (*SerbianHybridDisambiguator)(nil)
