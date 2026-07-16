package es

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

// SpanishSynthesizer ports synthesis.es.SpanishSynthesizer.
type SpanishSynthesizer struct {
	*synthesis.BaseSynthesizer
}

func NewSpanishSynthesizer(manual *synthesis.ManualSynthesizer) *SpanishSynthesizer {
	base := synthesis.NewBaseSynthesizer("es", manual)
	base.ResourceFileName = "/es/spanish_synth.dict"
	base.TagFileName = "/es/spanish_tags.txt"
	return &SpanishSynthesizer{BaseSynthesizer: base}
}

func (s *SpanishSynthesizer) Synthesize(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
	return s.BaseSynthesizer.Synthesize(token, posTag)
}
func (s *SpanishSynthesizer) SynthesizeRE(token *languagetool.AnalyzedToken, posTag string, re bool) ([]string, error) {
	return s.BaseSynthesizer.SynthesizeRE(token, posTag, re)
}

var _ synthesis.Synthesizer = (*SpanishSynthesizer)(nil)
