package en

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	entok "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers/en"
)

// EnglishSuppressMisspelledSuggestionsFilter ports
// org.languagetool.rules.en.EnglishSuppressMisspelledSuggestionsFilter
// (empty subclass of AbstractSuppressMisspelledSuggestionsFilter).
//
// Java abstract isMisspelled:
//
//	SpellingCheckRule spellerRule = language.getDefaultSpellingRule();
//	if (spellerRule == null) return false;
//	for (String token : language.getWordTokenizer().tokenize(s))
//	  if (spellerRule.isMisspelled(token)) return true;
//
// EN: getDefaultSpellingRule → MorfologikAmerican (when wired); else null → keep all.
// getWordTokenizer → EnglishWordTokenizer.
// Default IsMisspelled is FilterDictIsMisspelled (false when no WireEnglishFilterSpeller).
type EnglishSuppressMisspelledSuggestionsFilter struct {
	*rules.AbstractSuppressMisspelledSuggestionsFilter
}

func NewEnglishSuppressMisspelledSuggestionsFilter() *EnglishSuppressMisspelledSuggestionsFilter {
	wt := entok.NewEnglishWordTokenizer()
	return &EnglishSuppressMisspelledSuggestionsFilter{
		AbstractSuppressMisspelledSuggestionsFilter: &rules.AbstractSuppressMisspelledSuggestionsFilter{
			// Java: language.getDefaultSpellingRule().isMisspelled(token)
			IsMisspelled: FilterDictIsMisspelled,
			// Java: language.getWordTokenizer().tokenize(s) → EnglishWordTokenizer
			Tokenize: wt.Tokenize,
		},
	}
}

// SetIsMisspelled wires the speller hook (tests / host override).
func (f *EnglishSuppressMisspelledSuggestionsFilter) SetIsMisspelled(fn func(string) bool) {
	if f == nil || f.AbstractSuppressMisspelledSuggestionsFilter == nil {
		return
	}
	f.IsMisspelled = fn
}

// FilterSuggestions is a thin helper over AcceptRuleMatch (same filter path as Java).
// drops misspelled suggestions; empty + suppressMatch → drop match.
func (f *EnglishSuppressMisspelledSuggestionsFilter) FilterSuggestions(suggs []string, suppressMatch bool) (kept []string, keepMatch bool) {
	if f == nil || f.AbstractSuppressMisspelledSuggestionsFilter == nil {
		// Java: null SpellingCheckRule → isMisspelled false for all → keep all
		return suggs, true
	}
	m := rules.NewRuleMatch(rules.NewFakeRule("S"), nil, 0, 1, "msg")
	m.SetSuggestedReplacements(append([]string(nil), suggs...))
	sm := "true"
	if !suppressMatch {
		sm = "false"
	}
	out := f.AbstractSuppressMisspelledSuggestionsFilter.AcceptRuleMatch(m, map[string]string{"suppressMatch": sm})
	if out == nil {
		return nil, false
	}
	return out.GetSuggestedReplacements(), true
}

// AcceptRuleMatch ports AbstractSuppressMisspelledSuggestionsFilter.acceptRuleMatch.
func (f *EnglishSuppressMisspelledSuggestionsFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if f == nil || f.AbstractSuppressMisspelledSuggestionsFilter == nil {
		return nil
	}
	return f.AbstractSuppressMisspelledSuggestionsFilter.AcceptRuleMatch(match, arguments)
}
