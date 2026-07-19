package ca

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// CatalanSuppressMisspelledSuggestionsFilter ports
// org.languagetool.rules.ca.CatalanSuppressMisspelledSuggestionsFilter
// (overrides AbstractSuppressMisspelledSuggestionsFilter.isMisspelled).
//
// Java isMisspelled(s, language):
//  1. null default spelling rule → true (unlike generic abstract's false)
//  2. analyzeText(s); any token with ChunkTag "_incorrect_verb_" → true
//  3. else spellerRule.match(sentence).length > 0
//
// HasIncorrectVerb ports step 2 (chunker); nil skips that branch (no invent chunks).
// Without WireCatalanFilterSpeller and without IsMisspelled override → true (Java null speller).
type CatalanSuppressMisspelledSuggestionsFilter struct {
	*rules.AbstractSuppressMisspelledSuggestionsFilter
	// HasIncorrectVerb reports analyze+chunk tag "_incorrect_verb_" on suggestion text.
	HasIncorrectVerb func(s string) bool
}

// incorrectVerbChunk is Java ChunkTag("_incorrect_verb_").
const incorrectVerbChunk = "_incorrect_verb_"

func NewCatalanSuppressMisspelledSuggestionsFilter() *CatalanSuppressMisspelledSuggestionsFilter {
	f := &CatalanSuppressMisspelledSuggestionsFilter{
		AbstractSuppressMisspelledSuggestionsFilter: &rules.AbstractSuppressMisspelledSuggestionsFilter{},
	}
	// Wire Catalan isMisspelled into abstract AcceptRuleMatch path.
	f.AbstractSuppressMisspelledSuggestionsFilter.IsMisspelled = f.catalanIsMisspelled
	return f
}

// catalanIsMisspelled ports CatalanSuppressMisspelledSuggestionsFilter.isMisspelled.
func (f *CatalanSuppressMisspelledSuggestionsFilter) catalanIsMisspelled(s string) bool {
	if f == nil {
		return true
	}
	// Host override of abstract field after New still works if they SetIsMisspelled —
	// but default path uses this method only when field points here.
	// Tests may replace f.IsMisspelled entirely.

	// Java: SpellingCheckRule spellerRule = language.getDefaultSpellingRule(); if null return true
	if !FilterDictAvailable() {
		return true
	}
	// Java: hasIncorrectVerb chunk
	if f.HasIncorrectVerb != nil && f.HasIncorrectVerb(s) {
		return true
	}
	// Java: spellerRule.match(sentence).length > 0
	// Without full JLanguageTool analyze, probe whitespace tokens with FilterDict
	// (same resource as Catalan default Morfologik speller when wired).
	return caSpellerMatchHasHits(s)
}

// caSpellerMatchHasHits approximates spellerRule.match(analyzedSentence).length > 0
// using the wired filter dict: true if any non-empty token is unknown.
func caSpellerMatchHasHits(s string) bool {
	fields := strings.Fields(s)
	if len(fields) == 0 {
		// empty / whitespace-only: no letter tokens → Java may still match empty
		return FilterDictIsMisspelled(s)
	}
	for _, tok := range fields {
		if FilterDictIsMisspelled(tok) {
			return true
		}
	}
	return false
}

// SetIsMisspelled replaces the Catalan isMisspelled hook (tests).
func (f *CatalanSuppressMisspelledSuggestionsFilter) SetIsMisspelled(fn func(string) bool) {
	if f == nil || f.AbstractSuppressMisspelledSuggestionsFilter == nil {
		return
	}
	if fn == nil {
		f.IsMisspelled = f.catalanIsMisspelled
		return
	}
	f.IsMisspelled = fn
}

// FilterSuggestions is a test helper matching prior surface API.
func (f *CatalanSuppressMisspelledSuggestionsFilter) FilterSuggestions(suggs []string, suppressMatch bool) (kept []string, keepMatch bool) {
	var miss rules.MisspelledFunc
	if f != nil && f.AbstractSuppressMisspelledSuggestionsFilter != nil {
		miss = f.IsMisspelled
	}
	if miss == nil {
		// Should not happen after New; Catalan null-speller semantics → all misspelled
		miss = func(string) bool { return true }
	}
	for _, s := range suggs {
		if !miss(s) {
			kept = append(kept, s)
		}
	}
	if len(kept) == 0 && suppressMatch {
		return nil, false
	}
	return kept, true
}

// AcceptRuleMatch ports AbstractSuppressMisspelledSuggestionsFilter.acceptRuleMatch.
func (f *CatalanSuppressMisspelledSuggestionsFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if f == nil || f.AbstractSuppressMisspelledSuggestionsFilter == nil {
		return nil
	}
	return f.AbstractSuppressMisspelledSuggestionsFilter.AcceptRuleMatch(match, arguments)
}
