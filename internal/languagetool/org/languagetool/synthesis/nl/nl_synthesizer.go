package nl

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

// DutchSynthesizer ports synthesis.nl.DutchSynthesizer.
type DutchSynthesizer struct {
	*synthesis.BaseSynthesizer
}

func NewDutchSynthesizer(manual *synthesis.ManualSynthesizer) *DutchSynthesizer {
	base := synthesis.NewBaseSynthesizer("nl", manual)
	base.ResourceFileName = "/nl/nl_synth.dict"
	base.TagFileName = "/nl/nl_tags.txt"
	return &DutchSynthesizer{BaseSynthesizer: base}
}

func (s *DutchSynthesizer) Synthesize(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
	return s.BaseSynthesizer.Synthesize(token, posTag)
}
func (s *DutchSynthesizer) SynthesizeRE(token *languagetool.AnalyzedToken, posTag string, re bool) ([]string, error) {
	return s.BaseSynthesizer.SynthesizeRE(token, posTag, re)
}

var _ synthesis.Synthesizer = (*DutchSynthesizer)(nil)
