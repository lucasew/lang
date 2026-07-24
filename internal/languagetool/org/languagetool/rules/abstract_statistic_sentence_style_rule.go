package rules

import (
	"regexp"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// AbstractStatisticSentenceStyleRule ports
// org.languagetool.rules.AbstractStatisticSentenceStyleRule.
// ConditionFulfilled inspects one sentence part's non-whitespace tokens (Java List).
// Java ctor: CREATIVE_WRITING, Style; setDefaultOff when !defaultActive.
type AbstractStatisticSentenceStyleRule struct {
	ID          string
	Description string
	MinPercent  int
	// Denominator is 100.0 for percent (Java default).
	Denominator float64
	// ExcludeDirectSpeech mirrors Java excludeDirectSpeech(); default true when unset via Zero value false - set explicitly.
	ExcludeDirectSpeech bool
	// WithoutDirectSpeech when true excludes DS even at MinPercent 0 (Java withoutDirectSpeech).
	WithoutDirectSpeech bool
	// Category ports Rule.category (Java CREATIVE_WRITING).
	Category *Category
	// IssueType ports getLocQualityIssueType (Java Style).
	IssueType ITSIssueType
	// DefaultOff ports setDefaultOff (Java when defaultActive=false).
	DefaultOff bool
	// ConditionFulfilled returns a token that marks a hit for the sentence part, or nil.
	ConditionFulfilled func(tokens []*languagetool.AnalyzedTokenReadings) *languagetool.AnalyzedTokenReadings
	// LimitMessage builds the over-limit message.
	LimitMessage func(limit int, percent float64) string
}

// InitStatisticSentenceStyleMeta applies Java AbstractStatisticSentenceStyleRule ctor metadata.
func InitStatisticSentenceStyleMeta(r *AbstractStatisticSentenceStyleRule, messages map[string]string, defaultActive bool) {
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

func (r *AbstractStatisticSentenceStyleRule) GetCategory() *Category {
	if r == nil {
		return nil
	}
	return r.Category
}

func (r *AbstractStatisticSentenceStyleRule) GetLocQualityIssueType() ITSIssueType {
	if r == nil || r.IssueType == "" {
		return ITSStyle
	}
	return r.IssueType
}

func (r *AbstractStatisticSentenceStyleRule) IsDefaultOff() bool { return r != nil && r.DefaultOff }

// MinToCheckParagraph ports AbstractStatisticSentenceStyleRule.minToCheckParagraph (Java returns -1).
func (r *AbstractStatisticSentenceStyleRule) MinToCheckParagraph() int { return -1 }

// EstimateContextForSureMatch ports TextLevelRule (Java always -1).
func (r *AbstractStatisticSentenceStyleRule) EstimateContextForSureMatch() int { return -1 }

func (r *AbstractStatisticSentenceStyleRule) GetDescription() string {
	if r != nil && r.Description != "" {
		return r.Description
	}
	return ""
}

var (
	statOpeningQuotes = regexp.MustCompile(`^["“„»«]$`)
	statEndingQuotes  = regexp.MustCompile(`^["“”»«]$`)
	// Java: Pattern.compile("[,;.:?•!-–—]")
	// ASCII '-' between '!' and en-dash is a character-class range (U+0021…U+2013).
	// Twin bug-for-bug; • (U+2022) and em-dash sit outside that range and are listed.
	statMarksRE = regexp.MustCompile(`^[\x{0021}-\x{2013}\x{2014}•]$`)
)

// IsStatisticMark ports AbstractStatisticSentenceStyleRule.isMark.
func IsStatisticMark(token *languagetool.AnalyzedTokenReadings) bool {
	return token != nil && statMarksRE.MatchString(token.GetToken())
}

// IsStatisticOpeningQuote ports AbstractStatisticSentenceStyleRule.isOpeningQuote.
func IsStatisticOpeningQuote(token *languagetool.AnalyzedTokenReadings) bool {
	return token != nil && statOpeningQuotes.MatchString(token.GetToken())
}

func (r *AbstractStatisticSentenceStyleRule) GetID() string {
	if r != nil && r.ID != "" {
		return r.ID
	}
	return "STATISTIC_SENTENCE_STYLE"
}

func (r *AbstractStatisticSentenceStyleRule) denominator() float64 {
	if r != nil && r.Denominator > 0 {
		return r.Denominator
	}
	return 100.0
}

// MatchList ports TextLevelRule match for statistic sentence style.
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
	excludeDS := r.ExcludeDirectSpeech
	for _, sentence := range sentences {
		if sentence == nil {
			continue
		}
		tokens := sentence.GetTokensWithoutWhitespace()
		relevant := make([]*languagetool.AnalyzedTokenReadings, 0, len(tokens))
		isDirectSpeech := false
		isSentenceCount := false
		var foundToken *languagetool.AnalyzedTokenReadings
		for n := 1; n < len(tokens); n++ {
			token := tokens[n]
			if token == nil {
				continue
			}
			sToken := token.GetToken()
			if excludeDS && !isDirectSpeech && statOpeningQuotes.MatchString(sToken) &&
				n < len(tokens)-1 && tokens[n+1] != nil && !tokens[n+1].IsWhitespaceBefore() {
				isDirectSpeech = true
				if len(relevant) > 0 {
					isSentenceCount = true
					foundToken = r.ConditionFulfilled(relevant)
					if foundToken != nil {
						break
					}
				}
			} else if excludeDS && isDirectSpeech && statEndingQuotes.MatchString(sToken) &&
				n > 1 && !token.IsWhitespaceBefore() {
				isDirectSpeech = false
				relevant = relevant[:0]
			} else if (!isDirectSpeech || (r.MinPercent == 0 && !r.WithoutDirectSpeech)) && !token.IsWhitespace() {
				relevant = append(relevant, token)
			}
			if n == len(tokens)-1 && len(relevant) > 0 {
				isSentenceCount = true
				foundToken = r.ConditionFulfilled(relevant)
			}
		}
		if isSentenceCount {
			sentenceCount++
		}
		if foundToken != nil {
			hits = append(hits, hit{
				sent: sentence,
				from: foundToken.GetStartPos() + pos,
				to:   foundToken.GetEndPos() + pos,
			})
		}
		pos += sentence.GetCorrectedTextLength()
	}
	if sentenceCount == 0 {
		return nil
	}
	percent := float64(len(hits)) * r.denominator() / float64(sentenceCount)
	if percent <= float64(r.MinPercent) {
		return nil
	}
	msgFn := r.LimitMessage
	if msgFn == nil {
		msgFn = func(limit int, p float64) string { return "Sentence style limit exceeded" }
	}
	msg := msgFn(r.MinPercent, percent)
	var out []*RuleMatch
	for _, h := range hits {
		out = append(out, NewRuleMatch(r, h.sent, h.from, h.to, msg))
	}
	return out
}
