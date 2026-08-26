package pt

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// AdvancedSynthesizerFilter ports org.languagetool.rules.pt.AdvancedSynthesizerFilter
// (empty subclass of AbstractAdvancedSynthesizerFilter).
//
// Java uses language.getSynthesizer().synthesize(token, desiredPostag, true).
// Without a wired synthesizer, Accept fails closed (do not invent forms).
type AdvancedSynthesizerFilter struct {
	*rules.AbstractAdvancedSynthesizerFilter
}

var defaultSynth func(lemma, postag string) []string

// WireDefaultSynthesize installs the process-wide synthesizer for this filter
// (Java: Portuguese.getSynthesizer()).
func WireDefaultSynthesize(fn func(lemma, postag string) []string) {
	defaultSynth = fn
}

// ClearDefaultSynthesize clears the process-wide synthesizer (tests).
func ClearDefaultSynthesize() {
	defaultSynth = nil
}

func NewAdvancedSynthesizerFilter() *AdvancedSynthesizerFilter {
	f := &AdvancedSynthesizerFilter{
		AbstractAdvancedSynthesizerFilter: &rules.AbstractAdvancedSynthesizerFilter{
			// Java AdvancedSynthesizerFilter.isSuggestionException
			IsSuggestionException: func(token, desiredPostag string) bool {
				if (desiredPostag == "VMIP1P0" || desiredPostag == "VMIP2P0") && !strings.HasSuffix(token, "s") {
					return true
				}
				return false
			},
			// Java getCompositePostag + movePronounTag for postagReplace=keepPronoun
			GetCompositePostag: func(lemmaSelect, postagSelect, originalPostag, desiredPostag, postagReplace string) string {
				if postagReplace == "keepPronoun" {
					return movePronounTag(originalPostag, desiredPostag)
				}
				return rules.GetCompositePostag(lemmaSelect, postagSelect, originalPostag, desiredPostag, postagReplace)
			},
		},
	}
	if defaultSynth != nil {
		f.Synthesize = resolveDefaultSynth
	}
	return f
}

// movePronounTag ports Portuguese AdvancedSynthesizerFilter.movePronounTag.
func movePronounTag(sourceTag, destinationTag string) string {
	sourceTagParts := strings.Split(sourceTag, ":")
	newTag := destinationTag
	if len(sourceTagParts) == 2 {
		destinationTagParts := strings.Split(destinationTag, ":")
		if len(destinationTagParts) > 0 {
			newTag = destinationTagParts[0] + ":" + sourceTagParts[1]
		}
	}
	return newTag
}

func resolveDefaultSynth(lemma, postag string) []string {
	if defaultSynth == nil {
		return nil
	}
	return defaultSynth(lemma, postag)
}

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
