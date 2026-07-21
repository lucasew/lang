package spelling

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules"
)

// Message keys from MessagesBundle.properties (ResourceBundle.getString).
const (
	// SpellingMessage ports messages.getString("spelling").
	SpellingMessage = "Possible spelling mistake found."
	// DescSpelling ports messages.getString("desc_spelling").
	DescSpelling = "Possible spelling mistake"
	// DescSpellingShort ports messages.getString("desc_spelling_short").
	DescSpellingShort = "Spelling mistake"
	// DescSpellingNoSuggestions ports messages.getString("desc_spelling_no_suggestions").
	DescSpellingNoSuggestions = "Possible spelling mistake (without suggestions)"
)

// CreateWrongSplitMatch ports SpellingCheckRule.createWrongSplitMatch.
//
//	if (!ruleMatchesSoFar.isEmpty() && last.fromPos == prevPos) remove last;
//	new RuleMatch(this, sentence, prevPos, pos + coveredWord.length(),
//	              messages.getString("spelling"), messages.getString("desc_spelling_short"));
//	setType(UnknownWord);
//	setSuggestedReplacement((suggestion1 + " " + suggestion2).trim());
//
// rule is the concrete Rule (Java this). Positions are UTF-16 (Java String.length).
func CreateWrongSplitMatch(
	rule any,
	sentence *languagetool.AnalyzedSentence,
	ruleMatches *[]*rules.RuleMatch,
	pos int,
	coveredWord, suggestion1, suggestion2 string,
	prevPos int,
) *rules.RuleMatch {
	if ruleMatches != nil && len(*ruleMatches) > 0 {
		last := (*ruleMatches)[len(*ruleMatches)-1]
		if last != nil && last.GetFromPos() == prevPos {
			*ruleMatches = (*ruleMatches)[:len(*ruleMatches)-1]
		}
	}
	to := pos + javaStringLenSpell(coveredWord)
	m := NewSpellingRuleMatch(rule, sentence, prevPos, to)
	m.SetSuggestedReplacements([]string{strings.TrimSpace(suggestion1 + " " + suggestion2)})
	return m
}

// FilterDupes ports SpellingCheckRule.filterDupes (stream distinct, preserve first order).
func FilterDupes[T comparable](words []T) []T {
	if len(words) == 0 {
		return words
	}
	seen := make(map[T]struct{}, len(words))
	out := make([]T, 0, len(words))
	for _, w := range words {
		if _, ok := seen[w]; ok {
			continue
		}
		seen[w] = struct{}{}
		out = append(out, w)
	}
	return out
}

// NewSpellingRuleMatch ports common RuleMatch construction for spelling hits:
// messages.getString("spelling") + messages.getString("desc_spelling_short") + Type.UnknownWord.
func NewSpellingRuleMatch(rule any, sentence *languagetool.AnalyzedSentence, fromPos, toPos int) *rules.RuleMatch {
	m := rules.NewRuleMatch(rule, sentence, fromPos, toPos, SpellingMessage)
	m.SetShortMessage(DescSpellingShort)
	m.SetType(rules.RuleMatchTypeUnknownWord)
	return m
}
