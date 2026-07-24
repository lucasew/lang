package en

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// FindSuggestionsFilter ports org.languagetool.rules.en.FindSuggestionsFilter
// (extends AbstractFindSuggestionsFilter with EN MorfologikSpeller + EnglishTagger).
//
// Java:
//
//	DICT_FILENAME = "/en/english.dict"
//	ctor: lazy MorfologikSpeller if resourceExists
//	getTagger() → EnglishTagger.INSTANCE
//	getSpellingSuggestions(atr) → speller.findSimilarWords(atr.getToken())
//	getSynthesizer() default null — no replacements2 synth path
//
// Default hooks use process-wide WireEnglishFindSuggestionsSpeller (via
// WireEnglishFilterTagger on the same official english.dict) and FilterTagWord.
// Without a speller dict, Accept fails closed (Java would NPE on null speller).
type FindSuggestionsFilter struct {
	*rules.AbstractFindSuggestionsFilter
	// spellingOverride true when host/tests inject via SetSpellingSuggestions.
	spellingOverride bool
}

func NewFindSuggestionsFilter() *FindSuggestionsFilter {
	return &FindSuggestionsFilter{
		AbstractFindSuggestionsFilter: &rules.AbstractFindSuggestionsFilter{
			// Java: speller.findSimilarWords(atr.getToken())
			SpellingSuggestions: defaultENSpellingSuggestions,
			// Java: getTagger() → EnglishTagger.INSTANCE
			Tag: FilterTagWord,
			// Java: getSynthesizer() default null — do not invent English synth path
		},
	}
}

func defaultENSpellingSuggestions(atr *languagetool.AnalyzedTokenReadings) []string {
	if atr == nil || !FilterFindSuggestionsAvailable() {
		return nil
	}
	return FilterDictSuggest(atr.GetToken())
}

// SetSpellingSuggestions installs a host/test findSimilarWords-style hook
// (overrides process-wide MorfologikSpeller; used by twin tests).
func (f *FindSuggestionsFilter) SetSpellingSuggestions(fn func(atr *languagetool.AnalyzedTokenReadings) []string) {
	if f == nil || f.AbstractFindSuggestionsFilter == nil || fn == nil {
		return
	}
	f.spellingOverride = true
	f.SpellingSuggestions = fn
}

// AcceptRuleMatch ports AbstractFindSuggestionsFilter.acceptRuleMatch with EN fail-closed gate.
func (f *FindSuggestionsFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *rules.RuleMatch {
	if f == nil || f.AbstractFindSuggestionsFilter == nil || match == nil {
		return nil
	}
	// Java constructs MorfologikSpeller when resourceExists; without dict / override, fail-closed.
	if !f.spellingOverride && !FilterFindSuggestionsAvailable() {
		return nil
	}
	return f.AbstractFindSuggestionsFilter.AcceptRuleMatch(match, arguments, patternTokens, tokenPositions)
}
