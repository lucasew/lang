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
func (f *AdaptSuggestionsFilter) AcceptRuleMatch(match *RuleMatch, _ map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *RuleMatch {
	if match == nil {
		return nil
	}
	// Java: pattern rules adapt via Language.adaptSuggestion.
	suggs := match.GetSuggestedReplacements()
	if len(suggs) == 0 {
		return match
	}
	original := originalErrorStr(match)
	adapt := f.Adapt
	if adapt == nil {
		if code := ruleLanguageCode(match.Rule); code != "" {
			adapt = languagetool.AdaptSuggestionForLanguage(code)
		}
	}
	if adapt == nil {
		adapt = func(s, _ string) string { return s }
	}
	out := make([]string, len(suggs))
	for i, s := range suggs {
		out[i] = adapt(s, original)
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

func originalErrorStr(match *RuleMatch) string {
	if match == nil {
		return ""
	}
	// Java: match.getOriginalErrorStr() (may be empty until setOriginalErrorStr).
	if s := match.GetOriginalErrorStr(); s != "" {
		return s
	}
	if match.Sentence == nil {
		return ""
	}
	text := match.Sentence.GetText()
	if match.FromPos >= 0 && match.ToPos <= len(text) && match.FromPos < match.ToPos {
		return text[match.FromPos:match.ToPos]
	}
	return ""
}
