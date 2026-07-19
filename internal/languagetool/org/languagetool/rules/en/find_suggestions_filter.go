package en

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// FindSuggestionsFilter ports org.languagetool.rules.en.FindSuggestionsFilter
// (extends AbstractFindSuggestionsFilter with EN speller + EnglishTagger).
//
// Default hooks use process-wide FilterDict* / FilterSuggestionMatchesPostag
// from WireEnglishFilterSpeller / WireEnglishFilterTagger.
// Without a speller dict, Accept fails closed (Java always has MorfologikSpeller when resource exists).
type FindSuggestionsFilter struct {
	*rules.AbstractFindSuggestionsFilter
}

func NewFindSuggestionsFilter() *FindSuggestionsFilter {
	return &FindSuggestionsFilter{
		AbstractFindSuggestionsFilter: &rules.AbstractFindSuggestionsFilter{
			// Java: speller.findSimilarWords(atr.getToken())
			SpellingSuggestions: func(atr *languagetool.AnalyzedTokenReadings) []string {
				if atr == nil {
					return nil
				}
				return FilterDictSuggest(atr.GetToken())
			},
			// Java: getTagger().tag(suggestion).matchesPosTagRegex(desiredPostag)
			MatchesDesiredPostag: FilterSuggestionMatchesPostag,
		},
	}
}

// AcceptRuleMatch ports AbstractFindSuggestionsFilter.acceptRuleMatch with EN fail-closed gates.
func (f *FindSuggestionsFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *rules.RuleMatch {
	if f == nil || f.AbstractFindSuggestionsFilter == nil || match == nil {
		return nil
	}
	// Without speller dict, cannot produce spelling suggestions (fail-closed).
	if !FilterDictAvailable() {
		return nil
	}
	// Java diacriticsMode: if original already matches desiredPostag → drop.
	// AbstractFindSuggestionsFilter also handles Mode=diacritics; this pre-gate
	// matches prior EN register behavior using FilterOriginalMatchesPostag.
	if strings.EqualFold(arguments["Mode"], "diacritics") {
		desired := arguments["desiredPostag"]
		atr := resolveFindWordFrom(arguments["wordFrom"], match, patternTokens, tokenPositions)
		if atr != nil && FilterOriginalMatchesPostag(atr.GetToken(), desired) {
			return nil
		}
	}
	return f.AbstractFindSuggestionsFilter.AcceptRuleMatch(match, arguments, patternTokens, tokenPositions)
}

// resolveFindWordFrom mirrors AbstractFindSuggestionsFilter.resolveWordFrom for diacritics gate.
func resolveFindWordFrom(wordFrom string, match *rules.RuleMatch,
	patternTokens []*languagetool.AnalyzedTokenReadings, tokenPositions []int) *languagetool.AnalyzedTokenReadings {
	if wordFrom == "marker" || wordFrom == "inmarker" {
		for _, t := range patternTokens {
			if t != nil && t.GetStartPos() >= match.GetFromPos() && t.GetStartPos() < match.GetToPos() {
				return t
			}
		}
		if len(patternTokens) > 0 {
			return patternTokens[0]
		}
		return nil
	}
	// numeric wordFrom: 1-based pattern element index
	n := 0
	for _, r := range wordFrom {
		if r < '0' || r > '9' {
			return nil
		}
		n = n*10 + int(r-'0')
	}
	if n < 1 || n > len(patternTokens) {
		return nil
	}
	_ = tokenPositions
	return patternTokens[n-1]
}
