package filters

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

// ArabicAdvancedSynthesizerFilter ports org.languagetool.rules.ar.filters.ArabicAdvancedSynthesizerFilter
// (empty subclass of AbstractAdvancedSynthesizerFilter).
//
// Java: language.getSynthesizer().synthesize(token, desiredPostag, true)
// via Arabic → ArabicSynthesizer.INSTANCE.
// Go resolves synthesizer in order:
//  1. SetSynthesize / WireDefaultSynthesize (tests / host inject)
//  2. patterns.LanguageSynthesizer("ar") when official ArabicSynthesizer is registered
//  3. fail-closed (nil; do not invent surface forms)
type ArabicAdvancedSynthesizerFilter struct {
	*rules.AbstractAdvancedSynthesizerFilter
}

// defaultSynth is process-wide Arabic synthesizer bridge (tests / language load).
var defaultSynth func(lemma, postag string) []string

// WireDefaultSynthesize installs the process-wide synthesizer for this filter
// (Java: Arabic.getSynthesizer()).
func WireDefaultSynthesize(fn func(lemma, postag string) []string) {
	defaultSynth = fn
}

// ClearDefaultSynthesize clears the process-wide synthesizer (tests).
func ClearDefaultSynthesize() {
	defaultSynth = nil
}

// NewArabicAdvancedSynthesizerFilter constructs the filter and wires synth resolution.
func NewArabicAdvancedSynthesizerFilter() *ArabicAdvancedSynthesizerFilter {
	f := &ArabicAdvancedSynthesizerFilter{
		AbstractAdvancedSynthesizerFilter: &rules.AbstractAdvancedSynthesizerFilter{},
	}
	f.ensureSynth()
	return f
}

func resolveDefaultSynth(lemma, postag string) []string {
	if defaultSynth == nil {
		return nil
	}
	return defaultSynth(lemma, postag)
}

// synthesizeArabicFromRegistry ports language.getSynthesizer().synthesize(token, postag, true)
// via the official patterns.LanguageSynthesizer registry (ArabicSynthesizer when loaded).
func synthesizeArabicFromRegistry(lemma, postag string) []string {
	s := patterns.LanguageSynthesizer("ar")
	if s == nil {
		return nil
	}
	var posPtr *string
	if postag != "" {
		p := postag
		posPtr = &p
	}
	var lemPtr *string
	if lemma != "" {
		l := lemma
		lemPtr = &l
	}
	// Java: new AnalyzedToken("", desiredPostag, desiredLemma)
	tok := languagetool.NewAnalyzedToken("", posPtr, lemPtr)
	forms, err := s.SynthesizeRE(tok, postag, true)
	if err != nil || len(forms) == 0 {
		return nil
	}
	return forms
}

func (f *ArabicAdvancedSynthesizerFilter) ensureSynth() {
	if f == nil || f.AbstractAdvancedSynthesizerFilter == nil {
		return
	}
	if f.Synthesize != nil {
		return
	}
	if defaultSynth != nil {
		f.Synthesize = resolveDefaultSynth
		return
	}
	// Official path: Arabic synthesizer registered for pattern <match postag> / filters.
	if patterns.LanguageSynthesizer("ar") != nil {
		f.Synthesize = synthesizeArabicFromRegistry
	}
}

// SetSynthesize overrides the synthesizer hook (tests / host).
func (f *ArabicAdvancedSynthesizerFilter) SetSynthesize(fn func(lemma, postag string) []string) {
	if f == nil || f.AbstractAdvancedSynthesizerFilter == nil {
		return
	}
	f.Synthesize = fn
}

// AcceptRuleMatch ports AbstractAdvancedSynthesizerFilter.acceptRuleMatch.
func (f *ArabicAdvancedSynthesizerFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *rules.RuleMatch {
	if f == nil || f.AbstractAdvancedSynthesizerFilter == nil {
		return nil
	}
	f.ensureSynth()
	return f.AbstractAdvancedSynthesizerFilter.AcceptRuleMatch(match, arguments, patternTokens, tokenPositions)
}
