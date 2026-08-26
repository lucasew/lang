package rules

import (
	"regexp"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// AbstractStatisticStyleRule ports org.languagetool.rules.AbstractStatisticStyleRule.
// ConditionFulfilled returns end token index (>= i) when a hint starts at i, or -1.
// Java ctor: CREATIVE_WRITING, Style; setDefaultOff when !defaultActive.
type AbstractStatisticStyleRule struct {
	ID          string
	Description string
	MinPercent  int
	// Denominator: 100 percent, 1000 per-mill (Java NonSignificantVerbsRule).
	Denominator         float64
	ExcludeDirectSpeech bool
	WithoutDirectSpeech bool
	// Category ports Rule.category (Java CREATIVE_WRITING).
	Category *Category
	// IssueType ports getLocQualityIssueType (Java Style).
	IssueType ITSIssueType
	// DefaultOff ports setDefaultOff (Java when defaultActive=false).
	DefaultOff bool
	// ConditionFulfilled returns end index when hit, else -1.
	ConditionFulfilled func(tokens []*languagetool.AnalyzedTokenReadings, i int) int
	// SentenceConditionFulfilled when true emits getSentenceMessage immediately (rare).
	SentenceConditionFulfilled func(tokens []*languagetool.AnalyzedTokenReadings, n int) bool
	// SentenceMessage for sentenceCondition hits.
	SentenceMessage string
	// LimitMessage builds the over-limit message.
	LimitMessage func(limit int, percent float64) string
	// incorrectExamples / correctExamples port Rule.addExamplePair.
	incorrectExamples []IncorrectExample
	correctExamples   []CorrectExample
}

// AddExamplePair ports Rule.addExamplePair.
func (r *AbstractStatisticStyleRule) AddExamplePair(incorrect IncorrectExample, correct CorrectExample) {
	if r == nil {
		return
	}
	appendExamplePair(&r.incorrectExamples, &r.correctExamples, incorrect, correct)
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *AbstractStatisticStyleRule) GetIncorrectExamples() []IncorrectExample {
	if r == nil || len(r.incorrectExamples) == 0 {
		return nil
	}
	out := make([]IncorrectExample, len(r.incorrectExamples))
	copy(out, r.incorrectExamples)
	return out
}

// GetCorrectExamples ports Rule.getCorrectExamples.
func (r *AbstractStatisticStyleRule) GetCorrectExamples() []CorrectExample {
	if r == nil || len(r.correctExamples) == 0 {
		return nil
	}
	out := make([]CorrectExample, len(r.correctExamples))
	copy(out, r.correctExamples)
	return out
}

// InitStatisticStyleMeta applies Java AbstractStatisticStyleRule constructor metadata.
// defaultActive false → DefaultOff true (Java default for DE filler/style rules).
func InitStatisticStyleMeta(r *AbstractStatisticStyleRule, messages map[string]string, defaultActive bool) {
	if r == nil {
		return
	}
	if r.Category == nil {
		r.Category = CreativeWritingCategory(messages)
	}
	if r.IssueType == "" {
		r.IssueType = ITSStyle
	}
	if !defaultActive {
		r.DefaultOff = true
	}
}

func (r *AbstractStatisticStyleRule) GetCategory() *Category {
	if r == nil {
		return nil
	}
	return r.Category
}

func (r *AbstractStatisticStyleRule) GetLocQualityIssueType() ITSIssueType {
	if r == nil || r.IssueType == "" {
		return ITSStyle
	}
	return r.IssueType
}

func (r *AbstractStatisticStyleRule) IsDefaultOff() bool { return r != nil && r.DefaultOff }

func (r *AbstractStatisticStyleRule) GetDescription() string {
	if r != nil && r.Description != "" {
		return r.Description
	}
	return ""
}

var (
	statStyleOpenQ  = regexp.MustCompile(`^["“„»«]$`)
	statStyleCloseQ = regexp.MustCompile(`^["“”»«]$`)
)

func (r *AbstractStatisticStyleRule) GetID() string {
	if r != nil && r.ID != "" {
		return r.ID
	}
	return "STATISTIC_STYLE"
}

func (r *AbstractStatisticStyleRule) denominator() float64 {
	if r != nil && r.Denominator > 0 {
		return r.Denominator
	}
	return 100.0
}

// MatchList ports TextLevelRule match for statistic style (token hits / word count).
func (r *AbstractStatisticStyleRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*RuleMatch {
	if r == nil || r.ConditionFulfilled == nil {
		return nil
	}
	type hit struct {
		sent *languagetool.AnalyzedSentence
		from int
		to   int
	}
	var deferred []hit
	var immediate []*RuleMatch
	wordCount := 0
	pos := 0
	isDirectSpeech := false
	excludeDS := r.ExcludeDirectSpeech
	for _, sentence := range sentences {
		if sentence == nil {
			continue
		}
		tokens := sentence.GetTokensWithoutWhitespace()
		for n := 1; n < len(tokens); n++ {
			token := tokens[n]
			if token == nil {
				continue
			}
			sToken := token.GetToken()
			if excludeDS && !isDirectSpeech && statStyleOpenQ.MatchString(sToken) &&
				n < len(tokens)-1 && tokens[n+1] != nil && !tokens[n+1].IsWhitespaceBefore() {
				isDirectSpeech = true
			} else if excludeDS && isDirectSpeech && statStyleCloseQ.MatchString(sToken) &&
				n > 1 && !token.IsWhitespaceBefore() {
				isDirectSpeech = false
			} else if (!isDirectSpeech || (r.MinPercent == 0 && !r.WithoutDirectSpeech)) &&
				!token.IsWhitespace() && !token.IsNonWord() {
				wordCount++
				nEnd := r.ConditionFulfilled(tokens, n)
				if nEnd >= n {
					if r.SentenceConditionFulfilled != nil && r.SentenceConditionFulfilled(tokens, n) {
						msg := r.SentenceMessage
						if msg == "" {
							msg = "Style sentence hint"
						}
						immediate = append(immediate, NewRuleMatch(r, sentence, token.GetStartPos()+pos, token.GetEndPos()+pos, msg))
					} else if nEnd < len(tokens) && tokens[nEnd] != nil {
						deferred = append(deferred, hit{
							sent: sentence,
							from: token.GetStartPos() + pos,
							to:   tokens[nEnd].GetEndPos() + pos,
						})
					}
				}
			}
		}
		pos += sentence.GetCorrectedTextLength()
	}
	numMatches := len(deferred) + len(immediate)
	var percent float64
	if wordCount > 0 {
		percent = float64(numMatches) * r.denominator() / float64(wordCount)
	}
	out := immediate
	if percent > float64(r.MinPercent) {
		msgFn := r.LimitMessage
		if msgFn == nil {
			msgFn = func(limit int, p float64) string { return "Style hint exceeded limit" }
		}
		msg := msgFn(r.MinPercent, percent)
		for _, h := range deferred {
			out = append(out, NewRuleMatch(r, h.sent, h.from, h.to, msg))
		}
	}
	return out
}
