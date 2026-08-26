package ru

import (
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// RussianHybridDisambiguator ports org.languagetool.tagging.disambiguation.ru.RussianHybridDisambiguator:
// MultiWordChunker.getInstance("/ru/multiwords.txt") defaults, then XmlRuleDisambiguator(Russian) no global.
// Java order: disambiguator.disambiguate(chunker.disambiguate(input)) — multiwords then XML.
// Both stages are eagerly wired from official resources when present (Java final fields).
type RussianHybridDisambiguator struct {
	disambiguation.AbstractDisambiguator
	// Chunker is Java MultiWordChunker.getInstance("/ru/multiwords.txt") defaults.
	Chunker interface {
		Disambiguate(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
	}
	// Rules is Java XmlRuleDisambiguator (field name "disambiguator" in Java).
	Rules interface {
		Disambiguate(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
	}
}

// hybridOnce / hybridInstance port Java private static volatile INSTANCE + getInstance().
var (
	hybridOnce     sync.Once
	hybridInstance *RussianHybridDisambiguator
)

// GetInstance ports RussianHybridDisambiguator.getInstance() (double-check via sync.Once).
// Java createDefaultDisambiguator returns getInstance().
func GetInstance() *RussianHybridDisambiguator {
	hybridOnce.Do(func() {
		hybridInstance = NewRussianHybridDisambiguator()
	})
	return hybridInstance
}

// NewRussianHybridDisambiguator ports Java field init (private ctor path):
//
//	chunker = MultiWordChunker.getInstance("/ru/multiwords.txt"); // F,F,F defaults
//	disambiguator = new XmlRuleDisambiguator(Russian.getInstance()); // useGlobalDisambiguation=false
//
// Stages are wired when the same official resources Java loads are discoverable.
// Prefer GetInstance() for the process singleton (Java public API).
func NewRussianHybridDisambiguator() *RussianHybridDisambiguator {
	d := &RussianHybridDisambiguator{}
	if mw := RussianMultiWordChunker(); mw != nil {
		d.Chunker = mw
	}
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
	// Java RussianHybridDisambiguator:
	// return disambiguator.disambiguate(chunker.disambiguate(input));
	// i.e. multiword chunker first, then XML rules (inverted vs Polish).
	if d.Chunker != nil {
		out = d.Chunker.Disambiguate(out)
	}
	if d.Rules != nil {
		out = d.Rules.Disambiguate(out)
	}
	return out
}

var _ disambiguation.Disambiguator = (*RussianHybridDisambiguator)(nil)
