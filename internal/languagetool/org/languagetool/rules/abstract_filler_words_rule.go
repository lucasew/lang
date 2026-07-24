package rules

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// AbstractFillerWordsRule ports org.languagetool.rules.AbstractFillerWordsRule
// (extends AbstractStatisticStyleRule).
//
// Java DEFAULT_MIN_PERCENT = 8: report fillers only when their share of words
// exceeds that percentage (MatchList / text-level). MinPercent 0 reports every
// filler (and still counts words in direct speech for that mode).
// Direct speech is excluded when MinPercent > 0 (Java excludeDirectSpeech).
type AbstractFillerWordsRule struct {
	*AbstractStatisticStyleRule
	Messages map[string]string
	// ID / Description optional; copied onto AbstractStatisticStyleRule in Init.
	ID          string
	Description string
	ShortMsg    string
	Message     string
	FillerWords map[string]struct{}
	// IsException ports AbstractFillerWordsRule.isException.
	IsException func(tokens []*languagetool.AnalyzedTokenReadings, idx int) bool
}

// FillerWordsDefaultMinPercent is Java AbstractFillerWordsRule.DEFAULT_MIN_PERCENT.
const FillerWordsDefaultMinPercent = 8

// InitFillerWordsMeta applies Java AbstractFillerWordsRule constructor metadata and
// wires AbstractStatisticStyleRule for percentage-based MatchList.
// If AbstractStatisticStyleRule.MinPercent is still 0, sets FillerWordsDefaultMinPercent (8).
// Call SetMinPercent(0) after Init when a language needs “show all” (rare; not Java default).
func InitFillerWordsMeta(r *AbstractFillerWordsRule, messages map[string]string, defaultActive bool) {
	if r == nil {
		return
	}
	r.Messages = messages
	if r.AbstractStatisticStyleRule == nil {
		r.AbstractStatisticStyleRule = &AbstractStatisticStyleRule{}
	}
	stat := r.AbstractStatisticStyleRule
	if r.ID != "" {
		stat.ID = r.ID
	} else if stat.ID == "" {
		stat.ID = "FILLER_WORDS"
	}
	if r.Description != "" {
		stat.Description = r.Description
	} else if messages != nil {
		if s := messages["filler_words_rule_desc"]; s != "" {
			stat.Description = s
		}
	}
	if stat.MinPercent == 0 {
		stat.MinPercent = FillerWordsDefaultMinPercent
	}
	stat.ExcludeDirectSpeech = true
	stat.ConditionFulfilled = r.conditionFulfilled
	stat.SentenceConditionFulfilled = func(tokens []*languagetool.AnalyzedTokenReadings, n int) bool {
		return false
	}
	stat.LimitMessage = func(limit int, percent float64) string {
		if r.Message != "" {
			return r.Message
		}
		if messages != nil {
			if s := messages["filler_words_rule_msg"]; s != "" {
				return s
			}
		}
		return "Filler word"
	}
	InitStatisticStyleMeta(stat, messages, defaultActive)
}

// SetMinPercent sets the statistic threshold (0 = report all fillers).
func (r *AbstractFillerWordsRule) SetMinPercent(p int) {
	if r == nil {
		return
	}
	if r.AbstractStatisticStyleRule == nil {
		r.AbstractStatisticStyleRule = &AbstractStatisticStyleRule{}
	}
	r.AbstractStatisticStyleRule.MinPercent = p
}

// GetMinPercent returns the configured threshold.
func (r *AbstractFillerWordsRule) GetMinPercent() int {
	if r == nil || r.AbstractStatisticStyleRule == nil {
		return FillerWordsDefaultMinPercent
	}
	return r.AbstractStatisticStyleRule.MinPercent
}

func (r *AbstractFillerWordsRule) GetID() string {
	if r != nil && r.AbstractStatisticStyleRule != nil && r.AbstractStatisticStyleRule.ID != "" {
		return r.AbstractStatisticStyleRule.ID
	}
	return "FILLER_WORDS"
}

func (r *AbstractFillerWordsRule) GetDescription() string {
	if r != nil && r.Description != "" {
		return r.Description
	}
	if r != nil && r.AbstractStatisticStyleRule != nil {
		return r.AbstractStatisticStyleRule.GetDescription()
	}
	return ""
}

func (r *AbstractFillerWordsRule) isFiller(tok string) bool {
	if r == nil || r.FillerWords == nil {
		return false
	}
	_, ok := r.FillerWords[strings.ToLower(tok)]
	return ok
}

// conditionFulfilled ports AbstractFillerWordsRule.conditionFulfilled.
func (r *AbstractFillerWordsRule) conditionFulfilled(tokens []*languagetool.AnalyzedTokenReadings, n int) int {
	if r == nil || n < 0 || n >= len(tokens) || tokens[n] == nil {
		return -1
	}
	if r.isFiller(tokens[n].GetToken()) && (r.IsException == nil || !r.IsException(tokens, n)) {
		return n
	}
	return -1
}

// Match is a single-sentence convenience wrapping MatchList (Java TextLevelRule).
func (r *AbstractFillerWordsRule) Match(sentence *languagetool.AnalyzedSentence) []*RuleMatch {
	if r == nil || sentence == nil {
		return nil
	}
	return r.MatchList([]*languagetool.AnalyzedSentence{sentence})
}

// MatchList ports TextLevelRule.match via AbstractStatisticStyleRule.
func (r *AbstractFillerWordsRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*RuleMatch {
	if r == nil || r.AbstractStatisticStyleRule == nil {
		return nil
	}
	// Ensure condition is wired (InitFillerWordsMeta may not have been called).
	if r.ConditionFulfilled == nil {
		r.ConditionFulfilled = r.conditionFulfilled
	}
	return r.AbstractStatisticStyleRule.MatchList(sentences)
}
