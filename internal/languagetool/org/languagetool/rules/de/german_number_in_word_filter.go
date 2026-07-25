package de

import (
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// GermanNumberInWordFilter ports org.languagetool.rules.de.GermanNumberInWordFilter
// (extends AbstractNumberInWordFilter; isMisspelled/getSuggestions from
// GermanyGerman.getInstance().getDefaultSpellingRule() → GermanSpellerRule).
type GermanNumberInWordFilter struct {
	inner *rules.NumberInWordFilter
}

// Lazy default spelling rule — Java always has GermanyGerman.getDefaultSpellingRule()
// (GermanSpellerRule with official de/hunspell/de_DE.dict). Mirror EN static
// MorfologikAmericanSpellerRule pattern: construct once, no process-start inject gate.
var (
	germanNumberInWordSpellerOnce sync.Once
	germanNumberInWordSpeller     *GermanSpellerRule
)

// germanDefaultSpellingRule ports GermanyGerman.getInstance().getDefaultSpellingRule().
func germanDefaultSpellingRule() *GermanSpellerRule {
	germanNumberInWordSpellerOnce.Do(func() {
		// Java: createDefaultSpellingRule → new GermanSpellerRule(messages, GermanyGerman)
		r := NewGermanSpellerRule(nil)
		// Wire official /de/hunspell/de_DE.dict + getSpeller plain-text lists
		// (same resources Java GermanSpellerRule ctor / init loads).
		_ = r.InitFromDiscoveredResources()
		germanNumberInWordSpeller = r
	})
	return germanNumberInWordSpeller
}

// NewGermanNumberInWordFilter ports GermanNumberInWordFilter()
// super(GermanyGerman.getInstance()).
func NewGermanNumberInWordFilter() *GermanNumberInWordFilter {
	_ = germanDefaultSpellingRule()
	return &GermanNumberInWordFilter{
		inner: &rules.NumberInWordFilter{
			// Java: GermanyGerman.getInstance().getDefaultSpellingRule().isMisspelled / getSuggestions
			IsMisspelled:   germanNumberInWordIsMisspelled,
			GetSuggestions: germanNumberInWordGetSuggestions,
		},
	}
}

// germanNumberInWordIsMisspelled ports GermanNumberInWordFilter.isMisspelled.
func germanNumberInWordIsMisspelled(word string) bool {
	r := germanDefaultSpellingRule()
	if r == nil {
		// Java always constructs the speller; empty-dict IsMisspelled is fail-closed false.
		return false
	}
	return r.IsMisspelled(word)
}

// germanNumberInWordGetSuggestions ports GermanNumberInWordFilter.getSuggestions
// → GermanSpellerRule.getSuggestions (Go twin: Suggest).
func germanNumberInWordGetSuggestions(word string) []string {
	r := germanDefaultSpellingRule()
	if r == nil {
		return nil
	}
	return r.Suggest(word)
}

// Suggestions ports AbstractNumberInWordFilter.acceptRuleMatch candidate list.
func (f *GermanNumberInWordFilter) Suggestions(word string) []string {
	if f == nil || f.inner == nil {
		return nil
	}
	return f.inner.Suggestions(word)
}

// AcceptRuleMatch ports AbstractNumberInWordFilter.acceptRuleMatch.
func (f *GermanNumberInWordFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, patternTokenPos int,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *rules.RuleMatch {
	if f == nil || match == nil || f.inner == nil {
		return nil
	}
	return f.inner.AcceptRuleMatch(match, arguments, patternTokenPos, patternTokens, tokenPositions)
}
