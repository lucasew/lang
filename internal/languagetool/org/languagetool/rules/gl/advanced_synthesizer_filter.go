package gl

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

// AdvancedSynthesizerFilter ports org.languagetool.rules.gl.AdvancedSynthesizerFilter
// (empty subclass of AbstractAdvancedSynthesizerFilter).
//
// Java: language.getSynthesizer().synthesize(token, desiredPostag, true).
// Go resolves synthesizer in order:
//  1. SetSynthesize / WireDefaultSynthesize (tests / host inject)
//  2. patterns.LanguageSynthesizer("gl") when official GalicianSynthesizer is registered
//  3. fail-closed (nil; do not invent surface forms)
type AdvancedSynthesizerFilter struct {
	*rules.AbstractAdvancedSynthesizerFilter
}

// defaultSynth is process-wide Galician synthesizer bridge (tests / language load).
var defaultSynth func(lemma, postag string) []string

// WireDefaultSynthesize installs the process-wide synthesizer for this filter
// (Java: Galician.getSynthesizer()).
func WireDefaultSynthesize(fn func(lemma, postag string) []string) {
	defaultSynth = fn
}

// ClearDefaultSynthesize clears the process-wide synthesizer (tests).
func ClearDefaultSynthesize() {
	defaultSynth = nil
}

func NewAdvancedSynthesizerFilter() *AdvancedSynthesizerFilter {
	f := &AdvancedSynthesizerFilter{
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

// synthesizeGalicianFromRegistry ports language.getSynthesizer().synthesize(token, postag, true)
// via the official patterns.LanguageSynthesizer registry (GalicianSynthesizer when loaded).
func synthesizeGalicianFromRegistry(lemma, postag string) []string {
	s := patterns.LanguageSynthesizer("gl")
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

func (f *AdvancedSynthesizerFilter) ensureSynth() {
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
	// Official path: Galician synthesizer registered for pattern <match postag> / filters.
	if patterns.LanguageSynthesizer("gl") != nil {
		f.Synthesize = synthesizeGalicianFromRegistry
	}
}

// SetSynthesize overrides the synthesizer hook (tests / host).
func (f *AdvancedSynthesizerFilter) SetSynthesize(fn func(lemma, postag string) []string) {
	if f == nil || f.AbstractAdvancedSynthesizerFilter == nil {
		return
	}
	f.Synthesize = fn
}

// AcceptRuleMatch ports AbstractAdvancedSynthesizerFilter.acceptRuleMatch.
func (f *AdvancedSynthesizerFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *rules.RuleMatch {
	if f == nil || f.AbstractAdvancedSynthesizerFilter == nil {
		return nil
	}
	f.ensureSynth()
	return f.AbstractAdvancedSynthesizerFilter.AcceptRuleMatch(match, arguments, patternTokens, tokenPositions)
}
