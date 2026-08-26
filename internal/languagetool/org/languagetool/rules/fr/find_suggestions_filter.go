package fr

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/patterns"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// FindSuggestionsFilter ports org.languagetool.rules.fr.FindSuggestionsFilter.
//
// Wires AbstractFindSuggestionsFilter with French cleanSuggestion and multi-query
// spelling path (MakeWrong when tagged, strip/add trailing s).
//
// SpellingMatch ports morfologikRule.match(one-token sentence).suggestedReplacements.
// When nil, defaults to FilterDictSuggest if WireFrenchFilterSpeller is active;
// otherwise Accept fails closed (Java always has French.getDefaultSpellingRule()).
// Tag / Synthesize: FrenchTagger + FrenchSynthesizer (process-wide / LanguageSynthesizer).
type FindSuggestionsFilter struct {
	*rules.AbstractFindSuggestionsFilter
	// SpellingMatch returns suggested replacements for a one-token "sentence".
	// When nil and FilterDictAvailable, FilterDictSuggest is used.
	SpellingMatch func(word string) []string
}

// French cleanSuggestion: strip leading clitics/pronouns, keep first word.
var frCleanSuggestionRE = regexp.MustCompile(`(?i)^[smntl]'|^(nous|vous|le|la|les|me|te|se|leur|en|y) `)
var frEndsInVowel = regexp.MustCompile(`[aeioué]$`)

func NewFindSuggestionsFilter() *FindSuggestionsFilter {
	f := &FindSuggestionsFilter{
		AbstractFindSuggestionsFilter: &rules.AbstractFindSuggestionsFilter{
			// Java: getTagger() → FrenchTagger
			Tag: FilterTagWord,
			// Java: getSynthesizer().synthesize(at, desiredPostag, true)
			Synthesize: frenchFindSuggestionsSynthesize,
		},
	}
	f.CleanSuggestion = frCleanSuggestion
	// SpellingSuggestions built when SpellingMatch is resolved via EnsureSpellingHook
	return f
}

func frenchFindSuggestionsSynthesize(tok *languagetool.AnalyzedToken, postagRE string) []string {
	s := patterns.LanguageSynthesizer("fr")
	if s == nil || tok == nil {
		return nil
	}
	forms, err := s.SynthesizeRE(tok, postagRE, true)
	if err != nil || len(forms) == 0 {
		return nil
	}
	return forms
}

// defaultFRSpellingMatch ports morfologikRule.match → suggested replacements via FilterDict.
func defaultFRSpellingMatch(word string) []string {
	if !FilterDictAvailable() || word == "" {
		return nil
	}
	return FilterDictSuggest(word)
}

// EnsureSpellingHook installs SpellingSuggestions from SpellingMatch (French multi-query logic).
func (f *FindSuggestionsFilter) EnsureSpellingHook() {
	if f == nil || f.AbstractFindSuggestionsFilter == nil {
		return
	}
	f.SpellingSuggestions = func(atr *languagetool.AnalyzedTokenReadings) []string {
		return f.getSpellingSuggestions(atr)
	}
}

func frCleanSuggestion(s string) string {
	// remove pronouns before verbs
	output := frCleanSuggestionRE.ReplaceAllString(s, "")
	// check only first element
	parts := strings.Split(output, " ")
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}

// getSpellingSuggestions ports French FindSuggestionsFilter.getSpellingSuggestions.
func (f *FindSuggestionsFilter) getSpellingSuggestions(atr *languagetool.AnalyzedTokenReadings) []string {
	if f == nil || atr == nil {
		return nil
	}
	matchFn := f.SpellingMatch
	if matchFn == nil {
		matchFn = defaultFRSpellingMatch
	}
	w := atr.GetToken()
	if atr.IsTagged() {
		// Java FindSuggestionsFilter uses StringTools.makeWrong (not InterrogativeVerbFilter's private makeWrong).
		w = tools.MakeWrong(w)
	}
	var wordsToCheck []string
	wordsToCheck = append(wordsToCheck, w)
	if strings.HasSuffix(w, "s") && len(w) > 1 {
		wordsToCheck = append(wordsToCheck, w[:len(w)-1])
	}
	if frEndsInVowel.MatchString(w) {
		wordsToCheck = append(wordsToCheck, w+"s")
	}
	var suggestions []string
	seen := map[string]struct{}{}
	for _, word := range wordsToCheck {
		for _, s := range matchFn(word) {
			if _, ok := seen[s]; ok {
				continue
			}
			seen[s] = struct{}{}
			suggestions = append(suggestions, s)
		}
	}
	return suggestions
}

// AcceptRuleMatch ports AbstractFindSuggestionsFilter via French hooks.
func (f *FindSuggestionsFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *rules.RuleMatch {
	if f == nil || f.AbstractFindSuggestionsFilter == nil || match == nil {
		return nil
	}
	// Resolve speller: host SpellingMatch or process-wide French filter dict.
	if f.SpellingMatch == nil && !FilterDictAvailable() {
		// Java always has default spelling rule when language loads; without dict fail-closed.
		return nil
	}
	if f.SpellingSuggestions == nil {
		f.EnsureSpellingHook()
	}
	return f.AbstractFindSuggestionsFilter.AcceptRuleMatch(match, arguments, patternTokens, tokenPositions)
}
