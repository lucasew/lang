package ga

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// IrishHybridDisambiguator ports org.languagetool.tagging.disambiguation.ga.IrishHybridDisambiguator.
// Java: disambiguator.disambiguate(chunker.disambiguate(input)) — multiwords first, then XML.
// MultiWordChunker.getInstance("/ga/multiwords.txt") defaults, then
// XmlRuleDisambiguator(Irish.getInstance()) with useGlobalDisambiguation=false.
type IrishHybridDisambiguator struct {
	disambiguation.AbstractDisambiguator
	// Chunker is Java MultiWordChunker.getInstance("/ga/multiwords.txt") defaults.
	Chunker interface {
		Disambiguate(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
	}
	// Rules is Java XmlRuleDisambiguator (field name "disambiguator" in Java).
	Rules interface {
		Disambiguate(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
	}
}

// NewIrishHybridDisambiguator builds stages Java constructs as final fields.
// Chunker is wired from official /ga/multiwords.txt when discoverable
// (getInstance("/ga/multiwords.txt") defaults: F/F/F; no removePreviousTags; no ignoreSpelling).
// Rules is eagerly wired from official ga/disambiguation.xml when present
// (XmlRuleDisambiguator(Irish) — useGlobalDisambiguation=false).
func NewIrishHybridDisambiguator() *IrishHybridDisambiguator {
	d := &IrishHybridDisambiguator{}
	if mw := IrishMultiWordChunker(); mw != nil {
		d.Chunker = mw
	}
	if xml := IrishXmlRuleDisambiguator(); xml != nil {
		d.Rules = xml
	}
	return d
}

func (d *IrishHybridDisambiguator) Disambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if input == nil {
		return nil
	}
	out := input
	if d.Chunker != nil {
		out = d.Chunker.Disambiguate(out)
	}
	if d.Rules != nil {
		out = d.Rules.Disambiguate(out)
	}
	return out
}

var _ disambiguation.Disambiguator = (*IrishHybridDisambiguator)(nil)
