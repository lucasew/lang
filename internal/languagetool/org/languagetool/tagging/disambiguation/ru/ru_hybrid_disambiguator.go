package ru

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// RussianHybridDisambiguator ports org.languagetool.tagging.disambiguation.ru.RussianHybridDisambiguator:
// MultiWordChunker.getInstance("/ru/multiwords.txt") defaults, then XmlRuleDisambiguator(Russian) no global.
// Java order: disambiguator.disambiguate(chunker.disambiguate(input)) — multiwords then XML.
// Rules is eagerly wired from official ru/disambiguation.xml when present (Java final field).
type RussianHybridDisambiguator struct {
	disambiguation.AbstractDisambiguator
	Chunker interface {
		Disambiguate(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
	}
	// Rules is Java XmlRuleDisambiguator (field name "disambiguator" in Java).
	Rules interface {
		Disambiguate(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
	}
}

// NewRussianHybridDisambiguator ports Java field init: XmlRuleDisambiguator(Russian.getInstance())
// (useGlobalDisambiguation=false). Chunker is left for injectors / multiword load helpers
// (same pattern as Arabic hybrid: multiwords loaded by callers).
func NewRussianHybridDisambiguator() *RussianHybridDisambiguator {
	d := &RussianHybridDisambiguator{}
	if xml := RussianXmlRuleDisambiguator(); xml != nil {
		d.Rules = xml
	}
	return d
}

// NewRussianHybridDisambiguatorWithStages matches older call sites that pass stages.
func NewRussianHybridDisambiguatorWithStages(chunker, secondary disambiguation.Disambiguator) *RussianHybridDisambiguator {
	return &RussianHybridDisambiguator{Chunker: chunker, Rules: secondary}
}

func (d *RussianHybridDisambiguator) Disambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if d == nil || input == nil {
		return input
	}
	out := input
	// multiwords first, then XML (Java RussianHybridDisambiguator)
	if d.Chunker != nil {
		out = d.Chunker.Disambiguate(out)
	}
	if d.Rules != nil {
		out = d.Rules.Disambiguate(out)
	}
	return out
}

var _ disambiguation.Disambiguator = (*RussianHybridDisambiguator)(nil)
