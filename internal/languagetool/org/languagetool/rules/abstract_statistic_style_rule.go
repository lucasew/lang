package rules

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// AbstractStatisticStyleRule is a surface port of
// org.languagetool.rules.AbstractStatisticStyleRule for percentage-based style hints.
// Languages supply ConditionFulfilled; MinPercent 0 reports all hits.
type AbstractStatisticStyleRule struct {
	ID          string
	Description string
	MinPercent  int
	// ConditionFulfilled returns end token index (>= i) when a hint starts at i, or -1.
	ConditionFulfilled func(tokens []*languagetool.AnalyzedTokenReadings, i int) int
	// LimitMessage builds the over-limit message.
	LimitMessage func(limit int, percent float64) string
}

func (r *AbstractStatisticStyleRule) GetID() string {
	if r.ID != "" {
		return r.ID
	}
	return "STATISTIC_STYLE"
}

// MatchList counts condition hits across sentences and emits matches when percentage exceeds MinPercent.
func (r *AbstractStatisticStyleRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*RuleMatch {
	if r == nil || r.ConditionFulfilled == nil {
		return nil
	}
	type hit struct {
		sent *languagetool.AnalyzedSentence
		from int
		to   int
	}
	var hits []hit
	wordCount := 0
	pos := 0
	for _, s := range sentences {
		if s == nil {
			continue
		}
		tokens := s.GetTokensWithoutWhitespace()
		for i := 1; i < len(tokens); i++ {
			wordCount++
			end := r.ConditionFulfilled(tokens, i)
			if end >= i {
				hits = append(hits, hit{
					sent: s,
					from: pos + tokens[i].GetStartPos(),
					to:   pos + tokens[end].GetEndPos(),
				})
			}
		}
		pos += s.GetCorrectedTextLength()
	}
	if wordCount == 0 {
		return nil
	}
	pct := 100.0 * float64(len(hits)) / float64(wordCount)
	if r.MinPercent > 0 && pct <= float64(r.MinPercent) {
		return nil
	}
	// MinPercent 0 → show all
	msgFn := r.LimitMessage
	if msgFn == nil {
		msgFn = func(limit int, p float64) string {
			return "Style hint exceeded limit"
		}
	}
	msg := msgFn(r.MinPercent, pct)
	var out []*RuleMatch
	for _, h := range hits {
		out = append(out, NewRuleMatch(r, h.sent, h.from, h.to, msg))
	}
	return out
}
