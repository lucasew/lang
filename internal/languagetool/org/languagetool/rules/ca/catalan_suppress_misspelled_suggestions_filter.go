package ca

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// CatalanSuppressMisspelledSuggestionsFilter ports
// org.languagetool.rules.ca.CatalanSuppressMisspelledSuggestionsFilter
// (overrides AbstractSuppressMisspelledSuggestionsFilter.isMisspelled).
//
// Java isMisspelled(s, language) receives the full suggestion:
//  1. null default spelling rule → true (unlike generic abstract's false)
//  2. analyzeText(s); any token with ChunkTag "_incorrect_verb_" → true
//  3. else spellerRule.match(sentence).length > 0
//
// HasIncorrectVerb ports step 2 (chunker); nil skips that branch (no invent chunks).
// Step 3 without full JLanguageTool.match: WordTokenizer + FilterDict isMisspelled
// per token (parent abstract path; not invent strings.Fields).
type CatalanSuppressMisspelledSuggestionsFilter struct {
	*rules.AbstractSuppressMisspelledSuggestionsFilter
	// HasIncorrectVerb reports analyze+chunk tag "_incorrect_verb_" on full suggestion.
	HasIncorrectVerb func(s string) bool
}

// incorrectVerbChunk is Java ChunkTag("_incorrect_verb_").
const incorrectVerbChunk = "_incorrect_verb_"

func NewCatalanSuppressMisspelledSuggestionsFilter() *CatalanSuppressMisspelledSuggestionsFilter {
	f := &CatalanSuppressMisspelledSuggestionsFilter{
		AbstractSuppressMisspelledSuggestionsFilter: &rules.AbstractSuppressMisspelledSuggestionsFilter{},
	}
	// Java polymorphism: isMisspelled(full suggestion) — not per-token invent.
	f.IsMisspelledOverride = f.catalanIsMisspelled
	return f
}

// catalanIsMisspelled ports CatalanSuppressMisspelledSuggestionsFilter.isMisspelled
// on the full suggestion string.
func (f *CatalanSuppressMisspelledSuggestionsFilter) catalanIsMisspelled(s string) bool {
	if f == nil {
		return true
	}
	// Java: SpellingCheckRule spellerRule = language.getDefaultSpellingRule(); if null return true
	if !FilterDictAvailable() {
		return true
	}
	// Java: analyzeText(s) + ChunkTag("_incorrect_verb_") on any non-whitespace token
	if f.HasIncorrectVerb != nil && f.HasIncorrectVerb(s) {
		return true
	}
	// Java: spellerRule.match(sentence).length > 0
	// Incomplete without full analyze/match — twin of parent tokenize+isMisspelled:
	// language.getWordTokenizer().tokenize(s) + speller.isMisspelled(token).
	return caSpellerMatchHasHits(s, f)
}

// caSpellerMatchHasHits ports match length > 0 via WordTokenizer + dict isMisspelled.
// Do not invent strings.Fields (differs on punctuation/apostrophe tokens).
func caSpellerMatchHasHits(s string, f *CatalanSuppressMisspelledSuggestionsFilter) bool {
	var tokens []string
	if f != nil && f.AbstractSuppressMisspelledSuggestionsFilter != nil && f.Tokenize != nil {
		tokens = f.Tokenize(s)
	} else {
		tokens = tokenizers.NewWordTokenizer().Tokenize(s)
	}
	for _, tok := range tokens {
		if FilterDictIsMisspelled(tok) {
			return true
		}
	}
	return false
}

// SetIsMisspelled replaces the Catalan full-string isMisspelled override (tests).
func (f *CatalanSuppressMisspelledSuggestionsFilter) SetIsMisspelled(fn func(string) bool) {
	if f == nil || f.AbstractSuppressMisspelledSuggestionsFilter == nil {
		return
	}
	if fn == nil {
		f.IsMisspelledOverride = f.catalanIsMisspelled
		return
	}
	f.IsMisspelledOverride = fn
}

// FilterSuggestions is a test helper matching prior surface API.
func (f *CatalanSuppressMisspelledSuggestionsFilter) FilterSuggestions(suggs []string, suppressMatch bool) (kept []string, keepMatch bool) {
	miss := f.misspelledFull
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

func (f *CatalanSuppressMisspelledSuggestionsFilter) misspelledFull(s string) bool {
	if f == nil || f.AbstractSuppressMisspelledSuggestionsFilter == nil {
		return true
	}
	if f.IsMisspelledOverride != nil {
		return f.IsMisspelledOverride(s)
	}
	return f.catalanIsMisspelled(s)
}

// AcceptRuleMatch ports AbstractSuppressMisspelledSuggestionsFilter.acceptRuleMatch.
func (f *CatalanSuppressMisspelledSuggestionsFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if f == nil || f.AbstractSuppressMisspelledSuggestionsFilter == nil {
		return nil
	}
	return f.AbstractSuppressMisspelledSuggestionsFilter.AcceptRuleMatch(match, arguments)
}
