package de

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// AdvancedSynthesizerFilter ports org.languagetool.rules.de.AdvancedSynthesizerFilter
// (empty subclass of AbstractAdvancedSynthesizerFilter).
//
// Java uses language.getSynthesizer().synthesize(token, desiredPostag, true).
// Without a wired synthesizer, Accept fails closed (do not invent forms).
type AdvancedSynthesizerFilter struct {
	*rules.AbstractAdvancedSynthesizerFilter
}

// defaultSynth is process-wide German synthesizer bridge (tests / language load).
var defaultSynth func(lemma, postag string) []string

// WireDefaultSynthesize installs the process-wide synthesizer for this filter
// (Java: GermanyGerman.getSynthesizer()).
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
	// Java: language.getSynthesizer() (GermanSynthesizer). Process-wide wire overrides discovery.
	// Without either, fail-closed (do not invent forms).
	if defaultSynth != nil {
		f.Synthesize = resolveDefaultSynth
	} else if openDiscoveredGermanSynthesizer() != nil || openDiscoveredGermanSynthBase() != nil {
		f.Synthesize = synthesizeGermanRE
	}
	return f
}

func resolveDefaultSynth(lemma, postag string) []string {
	if defaultSynth == nil {
		return nil
	}
	return defaultSynth(lemma, postag)
}

// ensureSynth attaches process-wide or keeps instance hook.
func (f *AdvancedSynthesizerFilter) ensureSynth() {
	if f == nil || f.AbstractAdvancedSynthesizerFilter == nil {
		return
	}
	if f.Synthesize == nil && defaultSynth != nil {
		f.Synthesize = resolveDefaultSynth
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
