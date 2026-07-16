package xx

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// DemoDisambiguator ports org.languagetool.tagging.disambiguation.xx.DemoDisambiguator —
// identity disambiguator (copies input to output).
type DemoDisambiguator struct {
	disambiguation.AbstractDisambiguator
}

func NewDemoDisambiguator() *DemoDisambiguator {
	return &DemoDisambiguator{}
}

func (DemoDisambiguator) Disambiguate(input *languagetool.AnalyzedSentence) *languagetool.AnalyzedSentence {
	return input
}

var _ disambiguation.Disambiguator = DemoDisambiguator{}
