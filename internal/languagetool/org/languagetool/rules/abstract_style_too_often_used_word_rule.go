package rules

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// AbstractStyleTooOftenUsedWordRule is a surface stand-in for
// org.languagetool.rules.AbstractStyleTooOftenUsedWordRule.
// Languages supply IsCounted and LimitMessage; MinPercent 0 flags any repeated counted word.
type AbstractStyleTooOftenUsedWordRule struct {
	Messages    map[string]string
	ID          string
	Description string
	MinPercent  int // 0 = show all repeats of counted forms
	MinWords    int // Java default 100; use 0 in tests
	// IsCounted returns true if token should enter the frequency map.
	IsCounted func(tok *languagetool.AnalyzedTokenReadings, index int, tokens []*languagetool.AnalyzedTokenReadings) bool
	// Key maps a counted token to its map key (surface lower or lemma stand-in).
	Key func(tok *languagetool.AnalyzedTokenReadings) string
	// LimitMessage builds the match message for a key that exceeded the threshold.
	LimitMessage func(minPercent int) string
}

func (r *AbstractStyleTooOftenUsedWordRule) GetID() string {
	if r.ID != "" {
		return r.ID
	}
	return "TOO_OFTEN_USED_WORD"
}

func (r *AbstractStyleTooOftenUsedWordRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*RuleMatch {
	if r.IsCounted == nil || r.Key == nil {
		return nil
	}
	type occ struct {
		sent *languagetool.AnalyzedSentence
		from int
		to   int
		key  string
	}
	var all []occ
	counts := map[string]int{}
	pos := 0
	totalCounted := 0
	for _, s := range sentences {
		tokens := s.GetTokensWithoutWhitespace()
		for i := 1; i < len(tokens); i++ {
			if !r.IsCounted(tokens[i], i, tokens) {
				continue
			}
			k := r.Key(tokens[i])
			if k == "" {
				continue
			}
			counts[k]++
			totalCounted++
			all = append(all, occ{
				sent: s,
				from: pos + tokens[i].GetStartPos(),
				to:   pos + tokens[i].GetEndPos(),
				key:  k,
			})
		}
		pos += s.GetCorrectedTextLength()
	}
	minWords := r.MinWords
	if minWords <= 0 {
		// allow short texts when MinPercent is 0 (show-all)
		if r.MinPercent != 0 {
			return nil
		}
	} else if totalCounted < minWords {
		return nil
	}
	// which keys exceed threshold
	over := map[string]bool{}
	if r.MinPercent == 0 {
		for k, c := range counts {
			if c >= 2 {
				over[k] = true
			}
		}
	} else if totalCounted > 0 {
		for k, c := range counts {
			pct := 100.0 * float64(c) / float64(totalCounted)
			if pct > float64(r.MinPercent) {
				over[k] = true
			}
		}
	}
	if len(over) == 0 {
		return nil
	}
	msgFn := r.LimitMessage
	if msgFn == nil {
		msgFn = func(p int) string {
			return "Word used too often"
		}
	}
	msg := msgFn(r.MinPercent)
	var matches []*RuleMatch
	for _, o := range all {
		if !over[o.key] {
			continue
		}
		// only flag second+ occurrence when show-all
		rm := NewRuleMatch(r, o.sent, o.from, o.to, msg)
		rm.ShortMessage = "too often used"
		matches = append(matches, rm)
	}
	_ = tools.StartsWithUppercase
	_ = strings.ToLower
	return matches
}
