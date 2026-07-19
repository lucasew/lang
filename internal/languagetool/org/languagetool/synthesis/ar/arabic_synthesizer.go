package ar

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/synthesis"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
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

// InflectMafoulMutlq ports ArabicSynthesizer.inflectMafoulMutlq (static morph rule).
func InflectMafoulMutlq(word string) string {
	if word == "" {
		return word
	}
	teh := string(tools.ArabicTehMarbuta)
	if strings.HasSuffix(word, teh) {
		return word + string(tools.ArabicFathatan)
	}
	return word + string(tools.ArabicFathatan) + string(tools.ArabicAlef)
}

// InflectAdjectiveTanwinNasb ports ArabicSynthesizer.inflectAdjectiveTanwinNasb.
func InflectAdjectiveTanwinNasb(word string, feminin bool) string {
	if word == "" {
		return word
	}
	teh := string(tools.ArabicTehMarbuta)
	if feminin {
		if strings.HasSuffix(word, teh) {
			return word + string(tools.ArabicFathatan)
		}
		return word + teh + string(tools.ArabicFathatan)
	}
	// masculine: strip teh marbuta if present
	if strings.HasSuffix(word, teh) {
		return strings.TrimSuffix(word, teh)
	}
	return word + string(tools.ArabicFathatan) + string(tools.ArabicAlef)
}

// Instance methods match Java instance call sites (same as static helpers).
func (s *ArabicSynthesizer) InflectMafoulMutlq(word string) string {
	return InflectMafoulMutlq(word)
}

func (s *ArabicSynthesizer) InflectAdjectiveTanwinNasb(word string, feminin bool) string {
	return InflectAdjectiveTanwinNasb(word, feminin)
}

var _ synthesis.Synthesizer = (*ArabicSynthesizer)(nil)
