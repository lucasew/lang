package rules

import "github.com/lucasew/lang/internal/languagetool/org/languagetool"

func init() {
	// Wire faithful WordRepeatRule for languagetool.SimpleWordRepeatChecker demos/tests.
	languagetool.PreferredWordRepeatFactory = WordRepeatSentenceChecker
}

// WordRepeatSentenceChecker returns a SentenceChecker using WordRepeatRule
// (Java equalsIgnoreCase + name exceptions), not a soft case-sensitive invent.
func WordRepeatSentenceChecker(ruleID string) languagetool.SentenceChecker {
	r := NewWordRepeatRule(map[string]string{
		"repetition":            "Word repetition",
		"desc_repetition_short": "Word repetition",
		"desc_repetition":       "Word repetition (e.g. 'will will')",
	})
	if ruleID != "" {
		r.IDOverride = ruleID
	}
	return AsSentenceCheckerSimple(r.Match)
}
