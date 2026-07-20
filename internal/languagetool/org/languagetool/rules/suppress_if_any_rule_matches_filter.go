package rules

import (
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// MatchSpan is a simple from/to range for overlap checks.
type MatchSpan struct {
	From, To int
}

// SuppressIfAnyRuleMatchesFilter ports org.languagetool.rules.SuppressIfAnyRuleMatchesFilter.
// MatchesInSentence reports rule matches for a given rule ID in a rewritten sentence
// (Java: createDefaultJLanguageTool + rule.match).
type SuppressIfAnyRuleMatchesFilter struct {
	// MatchesInSentence returns match spans for ruleID in newSentence.
	MatchesInSentence func(ruleID, newSentence string) []MatchSpan
}

func NewSuppressIfAnyRuleMatchesFilter(fn func(ruleID, newSentence string) []MatchSpan) *SuppressIfAnyRuleMatchesFilter {
	return &SuppressIfAnyRuleMatchesFilter{MatchesInSentence: fn}
}

var (
	suppressAnyMu   sync.RWMutex
	defaultSuppress func(ruleID, newSentence string) []MatchSpan
)

// SetDefaultSuppressRuleMatcher wires re-check backend (Java JLanguageTool).
func SetDefaultSuppressRuleMatcher(fn func(ruleID, newSentence string) []MatchSpan) {
	suppressAnyMu.Lock()
	defer suppressAnyMu.Unlock()
	defaultSuppress = fn
}

// AcceptRuleMatch ports SuppressIfAnyRuleMatchesFilter.acceptRuleMatch.
// Without a re-check backend: fail-closed drop (cannot invent "no other rule matches").
func (f *SuppressIfAnyRuleMatchesFilter) AcceptRuleMatch(match *RuleMatch, arguments map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *RuleMatch {
	if match == nil {
		return nil
	}
	fn := f.MatchesInSentence
	if fn == nil {
		suppressAnyMu.RLock()
		fn = defaultSuppress
		suppressAnyMu.RUnlock()
	}
	if fn == nil {
		// Incomplete vs full LT re-check — do not keep match unfiltered (cheat).
		return nil
	}
	ruleIDs, ok := arguments["ruleIDs"]
	if !ok {
		panic("Missing key 'ruleIDs'")
	}
	sentence := ""
	if match.Sentence != nil {
		sentence = match.Sentence.GetText()
	}
	if (&SuppressIfAnyRuleMatchesFilter{MatchesInSentence: fn}).ShouldSuppress(
		sentence, match.GetFromPos(), match.GetToPos(), match.GetSuggestedReplacements(), ruleIDs) {
		return nil
	}
	return match
}

// ShouldSuppress is true if any replacement creates an overlapping match for any ruleIDs.
// Java: sentence.substring(0, fromPos) + replacement + sentence.substring(toPos)
// — fromPos/toPos are UTF-16 indices (Java String), not Go UTF-8 byte offsets.
func (f *SuppressIfAnyRuleMatchesFilter) ShouldSuppress(sentence string, fromPos, toPos int, replacements []string, ruleIDsCSV string) bool {
	if f.MatchesInSentence == nil {
		return false
	}
	ids := strings.Split(ruleIDsCSV, ",")
	textLen := utf16Len(sentence)
	for _, replacement := range replacements {
		// Java String.substring throws on invalid ranges; skip only invalid inputs
		// that would not appear from RuleMatch (keep fail-closed non-suppress).
		if fromPos < 0 || toPos > textLen || fromPos > toPos {
			continue
		}
		// Java: substring(0, fromPos) + replacement + substring(toPos)
		newSentence := utf16Substring(sentence, 0, fromPos) + replacement + utf16Substring(sentence, toPos, textLen)
		for _, id := range ids {
			id = strings.TrimSpace(id)
			for _, m := range f.MatchesInSentence(id, newSentence) {
				// overlap with original match range (Java logic)
				if (m.To >= fromPos && m.To <= toPos) || (toPos >= m.From && toPos <= m.To) {
					return true
				}
			}
		}
	}
	return false
}
