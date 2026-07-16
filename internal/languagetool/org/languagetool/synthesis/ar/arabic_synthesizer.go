package ar

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
)

const (
	ArabicSynthDict = "/ar/arabic_synth.dict"
	ArabicTagsFile  = "/ar/arabic_tags.txt"
)

// ArabicSynthesizer ports org.languagetool.synthesis.ar.ArabicSynthesizer.
type ArabicSynthesizer struct {
	*synthesis.BaseSynthesizer
}

func NewArabicSynthesizer(manual *synthesis.ManualSynthesizer) *ArabicSynthesizer {
	base := synthesis.NewBaseSynthesizer("ar", manual)
	base.ResourceFileName = ArabicSynthDict
	base.TagFileName = ArabicTagsFile
	return &ArabicSynthesizer{BaseSynthesizer: base}
}

// INSTANCE is the default shared synthesizer (no dict loaded until Manual is set).
var INSTANCE = NewArabicSynthesizer(nil)

func (s *ArabicSynthesizer) Synthesize(token *languagetool.AnalyzedToken, posTag string) ([]string, error) {
	return s.BaseSynthesizer.Synthesize(token, posTag)
}

func (s *ArabicSynthesizer) SynthesizeRE(token *languagetool.AnalyzedToken, posTag string, re bool) ([]string, error) {
	return s.BaseSynthesizer.SynthesizeRE(token, posTag, re)
}

var _ synthesis.Synthesizer = (*ArabicSynthesizer)(nil)
