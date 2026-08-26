package spelling

import (
	"fmt"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/suggestions"
)

// AddSuggestionsToRuleMatch ports SpellingCheckRule.addSuggestionsToRuleMatch.
//
//	if orderer != null && orderer.isMlAvailable():
//	  SuggestionsRanker → rank default, prepend user (no auto-correct with user)
//	  SuggestionsOrdererFeatureExtractor → computeFeatures (user candidates forbidden)
//	  else → order user then default
//	else: concat match's existing + user + default (no reranking)
func AddSuggestionsToRuleMatch(
	word string,
	userCandidatesList []*rules.SuggestedReplacement,
	candidatesList []*rules.SuggestedReplacement,
	orderer suggestions.SuggestionsOrderer,
	match *rules.RuleMatch,
) {
	if match == nil {
		return
	}
	sentence := match.Sentence
	userCandidates := replacementStrings(userCandidatesList)
	candidates := replacementStrings(candidatesList)
	startPos := match.GetFromPos()

	if orderer != nil && orderer.IsMlAvailable() {
		if ranker, ok := orderer.(suggestions.SuggestionsRanker); ok {
			// don't rank words from user dictionary; confidence null; add at start
			defaultSuggestions := ranker.OrderSuggestions(candidates, word, sentence, startPos)
			if len(defaultSuggestions) == 0 {
				// could not rank for some reason — leave match as-is
				return
			}
			if len(userCandidates) == 0 {
				match.SetAutoCorrect(ranker.ShouldAutoCorrect(defaultSuggestions))
				match.SetSuggestedReplacementObjects(defaultSuggestions)
			} else {
				combined := make([]*rules.SuggestedReplacement, 0, len(userCandidates)+len(defaultSuggestions))
				for _, w := range userCandidates {
					// confidence is null
					combined = append(combined, rules.NewSuggestedReplacement(w))
				}
				combined = append(combined, defaultSuggestions...)
				match.SetSuggestedReplacementObjects(combined)
				// no auto correct when words from personal dictionaries are included
				match.SetAutoCorrect(false)
			}
			return
		}
		if fe, ok := orderer.(*suggestions.SuggestionsOrdererFeatureExtractor); ok {
			// disable user suggestions here
			if len(userCandidates) != 0 {
				panic(fmt.Errorf(
					"SuggestionsOrdererFeatureExtractor does not support suggestions from personal dictionaries at the moment"))
			}
			ordered, feats := fe.ComputeFeatures(candidates, word, sentence, startPos)
			match.SetSuggestedReplacementObjects(ordered)
			match.SetFeatures(feats)
			return
		}
		// generic ML orderer: order user then default
		combined := make([]*rules.SuggestedReplacement, 0, len(userCandidates)+len(candidates))
		combined = append(combined, orderer.OrderSuggestions(userCandidates, word, sentence, startPos)...)
		combined = append(combined, orderer.OrderSuggestions(candidates, word, sentence, startPos)...)
		match.SetSuggestedReplacementObjects(combined)
		return
	}

	// no reranking
	combined := make([]*rules.SuggestedReplacement, 0,
		len(match.GetSuggestedReplacementObjects())+len(userCandidatesList)+len(candidatesList))
	combined = append(combined, match.GetSuggestedReplacementObjects()...)
	combined = append(combined, userCandidatesList...)
	combined = append(combined, candidatesList...)
	match.SetSuggestedReplacementObjects(combined)
}

// AddSuggestionsToRuleMatchStrings is a convenience for string candidate lists.
func AddSuggestionsToRuleMatchStrings(
	word string,
	userCandidates, candidates []string,
	orderer suggestions.SuggestionsOrderer,
	match *rules.RuleMatch,
) {
	AddSuggestionsToRuleMatch(word,
		rules.ConvertSuggestions(userCandidates),
		rules.ConvertSuggestions(candidates),
		orderer, match)
}

func replacementStrings(list []*rules.SuggestedReplacement) []string {
	if len(list) == 0 {
		return nil
	}
	out := make([]string, 0, len(list))
	for _, s := range list {
		if s != nil {
			out = append(out, s.GetReplacement())
		}
	}
	return out
}
