package rules

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// AbstractStatisticSentenceStyleRule ports
// org.languagetool.rules.AbstractStatisticSentenceStyleRule at sentence granularity.
// ConditionFulfilled returns a token that marks a hit for the sentence, or nil.
type AbstractStatisticSentenceStyleRule struct {
	ID          string
	Description string
	MinPercent  int
	// ConditionFulfilled inspects one sentence's non-whitespace tokens.
	ConditionFulfilled func(tokens []*languagetool.AnalyzedTokenReadings) *languagetool.AnalyzedTokenReadings
	// LimitMessage builds the over-limit message.
	LimitMessage func(limit int, percent float64) string
}

func (r *AbstractStatisticSentenceStyleRule) GetID() string {
	if r.ID != "" {
		return r.ID
	}
	return "STATISTIC_SENTENCE_STYLE"
}

// MatchList counts sentences with hits and emits matches when percentage exceeds MinPercent.
func (r *AbstractStatisticSentenceStyleRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*RuleMatch {
	if r == nil || r.ConditionFulfilled == nil {
		return nil
	}
	type hit struct {
		sent *languagetool.AnalyzedSentence
		from int
		to   int
	}
	var hits []hit
	sentenceCount := 0
	pos := 0
	for _, s := range sentences {
		if s == nil {
			continue
		}
		sentenceCount++
		tokens := s.GetTokensWithoutWhitespace()
		if tok := r.ConditionFulfilled(tokens); tok != nil {
			hits = append(hits, hit{
				sent: s,
				from: pos + tok.GetStartPos(),
				to:   pos + tok.GetEndPos(),
			})
		}
		pos += s.GetCorrectedTextLength()
	}
	if sentenceCount == 0 {
		return nil
	}
	pct := 100.0 * float64(len(hits)) / float64(sentenceCount)
	if r.MinPercent > 0 && pct <= float64(r.MinPercent) {
		return nil
	}
	msgFn := r.LimitMessage
	if msgFn == nil {
		msgFn = func(limit int, p float64) string { return "Sentence style limit exceeded" }
	}
	msg := msgFn(r.MinPercent, pct)
	var out []*RuleMatch
	for _, h := range hits {
		out = append(out, NewRuleMatch(r, h.sent, h.from, h.to, msg))
	}
	return out
}
