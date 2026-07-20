package es

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
)

// FindSuggestionsFilter ports org.languagetool.rules.es.FindSuggestionsFilter
// (extends AbstractFindSuggestionsFilter with MorfologikSpeller + SpanishTagger).
//
// Default SpellingSuggestions uses process-wide WireSpanishFilterSpeller /
// FilterDictSuggest (Java: speller.findSimilarWords; resource /es/es-ES.dict).
// Without a dict, Accept fails closed unless SetSpellingFromSimilarWords overrides.
// Tag / Synthesize: SpanishTagger + SpanishSynthesizer via process-wide hooks.
type FindSuggestionsFilter struct {
	*rules.AbstractFindSuggestionsFilter
	// spellingOverride true after SetSpellingFromSimilarWords (tests/host inject).
	spellingOverride bool
}

func NewFindSuggestionsFilter() *FindSuggestionsFilter {
	return &FindSuggestionsFilter{
		AbstractFindSuggestionsFilter: &rules.AbstractFindSuggestionsFilter{
			// Java: speller.findSimilarWords(atr.getToken())
			SpellingSuggestions: defaultESSpellingSuggestions,
			// Java: getTagger() → SpanishTagger
			Tag: FilterTagWord,
			// Java: getSynthesizer().synthesize(at, desiredPostag, true)
			Synthesize: spanishFindSuggestionsSynthesize,
		},
	}
}

func spanishFindSuggestionsSynthesize(tok *languagetool.AnalyzedToken, postagRE string) []string {
	s := patterns.LanguageSynthesizer("es")
	if s == nil || tok == nil {
		return nil
	}
	forms, err := s.SynthesizeRE(tok, postagRE, true)
	if err != nil || len(forms) == 0 {
		return nil
	}
	return forms
}

func defaultESSpellingSuggestions(atr *languagetool.AnalyzedTokenReadings) []string {
	if atr == nil || !FilterDictAvailable() {
		return nil
	}
	return FilterDictSuggest(atr.GetToken())
}

// AcceptRuleMatch ports AbstractFindSuggestionsFilter.acceptRuleMatch with ES fail-closed gate.
func (f *FindSuggestionsFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *rules.RuleMatch {
	if f == nil || f.AbstractFindSuggestionsFilter == nil || match == nil {
		return nil
	}
	// Java always constructs MorfologikSpeller when resource exists; without dict / override, fail-closed.
	if !f.spellingOverride && !FilterDictAvailable() {
		return nil
	}
	return f.AbstractFindSuggestionsFilter.AcceptRuleMatch(match, arguments, patternTokens, tokenPositions)
}

// SetSpellingFromSimilarWords installs SpellingSuggestions from a findSimilarWords-style hook
// (tests / host override of MorfologikSpeller.findSimilarWords).
func (f *FindSuggestionsFilter) SetSpellingFromSimilarWords(findSimilar func(token string) []string) {
	if f == nil || f.AbstractFindSuggestionsFilter == nil || findSimilar == nil {
		return
	}
	f.spellingOverride = true
	f.SpellingSuggestions = func(atr *languagetool.AnalyzedTokenReadings) []string {
		if atr == nil {
			return nil
		}
		return findSimilar(atr.GetToken())
	}
}
