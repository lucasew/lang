package pt

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// sentenceStep is a Disambiguate-capable stage (MultiWordChunker / XmlRuleDisambiguator).
type sentenceStep interface {
	Disambiguate(*languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence
}

// PortugueseHybridDisambiguator ports
// org.languagetool.tagging.disambiguation.pt.PortugueseHybridDisambiguator:
//
//	chunker = MultiWordChunker.getInstance("/pt/multiwords.txt", true, true, true);
//	chunkerGlobal = MultiWordChunker.getInstance("/spelling_global.txt", false, true, true, "NPCN000");
//	// ctor: chunkerGlobal.setIgnoreSpelling(true);
//	//       disambiguator = new XmlRuleDisambiguator(lang, true);
//	//       chunker.setRemovePreviousTags(true);
//	//       chunker.setIgnoreSpelling(true);
//
// Java disambiguate order (nested call):
//
//	return disambiguator.disambiguate(
//	         chunker.disambiguate(
//	           chunkerGlobal.disambiguate(input, checkCanceled), checkCanceled), checkCanceled);
//
// i.e. spelling_global FIRST, multiwords SECOND (removePreviousTags + ignoreSpelling),
// XmlRuleDisambiguator(lang, true) THIRD. Both chunkers ignore spelling; global has no
// setRemovePreviousTags (flattening of global angles is done by multiwords after).
// Stages are eagerly wired from the same official resources Java loads.
type PortugueseHybridDisambiguator struct {
	disambiguation.AbstractDisambiguator
	// GlobalChunker is Java chunkerGlobal (spelling_global + DefaultTag NPCN000).
	GlobalChunker sentenceStep
	// Chunker is Java chunker (/pt/multiwords.txt + removePreviousTags + ignoreSpelling).
	Chunker sentenceStep
	// Rules is Java disambiguator (XmlRuleDisambiguator(lang, true)).
	Rules sentenceStep
}

// NewPortugueseHybridDisambiguator ports Java field init + constructor side effects:
//
//	chunkerGlobal = MultiWordChunker.getInstance("/spelling_global.txt", false, true, true, "NPCN000");
//	chunkerGlobal.setIgnoreSpelling(true); // no setRemovePreviousTags
//	chunker = MultiWordChunker.getInstance("/pt/multiwords.txt", true, true, true);
//	chunker.setRemovePreviousTags(true);
//	chunker.setIgnoreSpelling(true);
//	disambiguator = new XmlRuleDisambiguator(lang, true); // useGlobalDisambiguation=true
//
// Stages are wired when the same official resources Java loads are discoverable.
func NewPortugueseHybridDisambiguator() *PortugueseHybridDisambiguator {
	d := &PortugueseHybridDisambiguator{}
	if g := PortugueseGlobalChunker(); g != nil {
		d.GlobalChunker = g
	}
	if mw := PortugueseMultiWordChunker(); mw != nil {
		d.Chunker = mw
	}
	if r := PortugueseXmlRuleDisambiguator(); r != nil {
		d.Rules = r
	}
	return d
}

func (d *PortugueseHybridDisambiguator) Disambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	if d == nil || input == nil {
		return input
	}
	out := input
	// Java PortugueseHybridDisambiguator:
	// return disambiguator.disambiguate(chunker.disambiguate(chunkerGlobal.disambiguate(...)));
	// i.e. global → multiwords → XML.
	if d.GlobalChunker != nil {
		out = d.GlobalChunker.Disambiguate(out)
	}
	if d.Chunker != nil {
		out = d.Chunker.Disambiguate(out)
	}
	if d.Rules != nil {
		out = d.Rules.Disambiguate(out)
	}
	return out
}

var _ disambiguation.Disambiguator = (*PortugueseHybridDisambiguator)(nil)
