package rules

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// AdaptSuggestionsFilter ports org.languagetool.rules.AdaptSuggestionsFilter.
// Adapt maps (replacement, originalError) → adjusted suggestion (Java: Language.adaptSuggestion).
// When Adapt is nil, uses languagetool.AdaptSuggestionForLanguage(rule language code) if set;
// otherwise identity (Java languages without override).
type AdaptSuggestionsFilter struct {
	Adapt func(replacement, originalError string) string
}

func NewAdaptSuggestionsFilter(adapt func(string, string) string) *AdaptSuggestionsFilter {
	return &AdaptSuggestionsFilter{Adapt: adapt}
}

// AcceptRuleMatch ports AdaptSuggestionsFilter.acceptRuleMatch.
// Java only adapts when rule instanceof AbstractPatternRule; otherwise returns match.
// Each replacement: match.setOriginalErrorStr() then lang.adaptSuggestion(replacement, getOriginalErrorStr()).
func (f *AdaptSuggestionsFilter) AcceptRuleMatch(match *RuleMatch, _ map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *RuleMatch {
	if match == nil {
		return nil
	}
	// Java: if (!(rule instanceof AbstractPatternRule)) return match;
	// Go: adapt only when language code is resolvable (PatternRule-like) or Adapt is set.
	suggs := match.GetSuggestedReplacements()
	if len(suggs) == 0 {
		return match
	}
	// Java: only AbstractPatternRule has a Language; else return match.
	code := ruleLanguageCode(match.Rule)
	if f.Adapt == nil && code == "" {
		return match
	}
	adapt := f.Adapt
	if adapt == nil {
		adapt = languagetool.AdaptSuggestionForLanguage(code)
	}
	if adapt == nil {
		// Java Language.adaptSuggestion default returns s unchanged.
		adapt = func(s, _ string) string { return s }
	}
	out := make([]string, len(suggs))
	for i, s := range suggs {
		// Java loop body: setOriginalErrorStr() then adaptSuggestion(replacement, getOriginalErrorStr()).
		match.SetOriginalErrorStr()
		out[i] = adapt(s, match.GetOriginalErrorStr())
	}
	match.SetSuggestedReplacements(out)
	return match
}

// ruleLanguageCode reads LanguageCode from PatternRule-like rules (Java getLanguage().getShortCode).
func ruleLanguageCode(rule any) string {
	if rule == nil {
		return ""
	}
	if r, ok := rule.(interface{ GetLanguageCode() string }); ok {
		return r.GetLanguageCode()
	}
	return ""
}

// MapSuggestions rewrites each suggestion using Adapt (unit tests).
func (f *AdaptSuggestionsFilter) MapSuggestions(suggs []string, originalError string) []string {
	adapt := f.Adapt
	if adapt == nil {
		adapt = func(s, _ string) string { return s }
	}
	out := make([]string, len(suggs))
	for i, s := range suggs {
		out[i] = adapt(s, originalError)
	}
	return out
}

// originalErrorStr is kept for unit tests that map suggestions with an explicit original.
// Production AcceptRuleMatch uses RuleMatch.SetOriginalErrorStr (UTF-16 span) like Java.
func originalErrorStr(match *RuleMatch) string {
	if match == nil {
		return ""
	}
	if s := match.GetOriginalErrorStr(); s != "" {
		return s
	}
	match.SetOriginalErrorStr()
	return match.GetOriginalErrorStr()
}
