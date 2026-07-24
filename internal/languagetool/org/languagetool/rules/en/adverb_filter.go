package en

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// AdverbFilter ports org.languagetool.rules.en.AdverbFilter.
// Filter that maps suggestion from adverb to adjective.
// Java @since 4.9
type AdverbFilter struct{}

func NewAdverbFilter() *AdverbFilter {
	return &AdverbFilter{}
}

// AcceptRuleMatch ports AdverbFilter.acceptRuleMatch.
// Java:
//
//	String adverb = arguments.get("adverb");
//	String noun = arguments.get("noun");
//	String adjective = adverb2Adj.get(adverb);
//	if (adjective != null && !adjective.equals(adverb)) {
//	  match.setSuggestedReplacement(adjective + " " + noun);
//	}
//	return match;
func (f *AdverbFilter) AcceptRuleMatch(match *rules.RuleMatch, arguments map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *rules.RuleMatch {
	if f == nil || match == nil {
		return nil
	}
	adverb := arguments["adverb"]
	noun := arguments["noun"]
	adjective, ok := adverb2Adj[adverb]
	if ok && adjective != adverb {
		// we can't simply cut off "ly" because of cases like "simply" -> "simple" etc.
		match.SetSuggestedReplacement(adjective + " " + noun)
	}
	return match
}
