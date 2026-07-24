package rules

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// AbstractStyleTooOftenUsedWordRule ports
// org.languagetool.rules.AbstractStyleTooOftenUsedWordRule (text-level; default off).
// MinWordCount is Java MIN_WORD_COUNT (100). MinPercent 0 is not Java (tests only).
// Java ctor: CREATIVE_WRITING, Style; setDefaultOff when !defaultActive.
type AbstractStyleTooOftenUsedWordRule struct {
	Messages     map[string]string
	ID           string
	Description  string
	MinPercent   int // threshold percent; Java defaults 5 for DE
	MinWordCount int // Java 100; 0 disables gate for twin tests
	// WithoutDirectSpeech when true excludes quoted spans (Java withoutDirectSpeech).
	WithoutDirectSpeech bool
	// Category ports Rule.category (Java CREATIVE_WRITING).
	Category *Category
	// IssueType ports getLocQualityIssueType (Java Style).
	IssueType ITSIssueType
	// DefaultOff ports setDefaultOff (Java when defaultActive=false).
	DefaultOff bool
	// IsToCountedWord ports isToCountedWord.
	IsToCountedWord func(tok *languagetool.AnalyzedTokenReadings) bool
	// IsException ports isException(tokens, n).
	IsException func(tokens []*languagetool.AnalyzedTokenReadings, n int) bool
	// ToAddedLemma ports toAddedLemma (nil → skip token).
	ToAddedLemma func(tok *languagetool.AnalyzedTokenReadings) string
	// LimitMessage ports getLimitMessage(minPercent).
	LimitMessage func(minPercent int) string
}

// InitStyleTooOftenUsedWordMeta applies Java AbstractStyleTooOftenUsedWordRule ctor metadata.
func InitStyleTooOftenUsedWordMeta(r *AbstractStyleTooOftenUsedWordRule, messages map[string]string, defaultActive bool) {
	if r == nil {
		return
	}
	r.Messages = messages
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

func (r *AbstractStyleTooOftenUsedWordRule) GetCategory() *Category {
	if r == nil {
		return nil
	}
	return r.Category
}

func (r *AbstractStyleTooOftenUsedWordRule) GetLocQualityIssueType() ITSIssueType {
	if r == nil || r.IssueType == "" {
		return ITSStyle
	}
	return r.IssueType
}

func (r *AbstractStyleTooOftenUsedWordRule) IsDefaultOff() bool { return r != nil && r.DefaultOff }

var (
	styleTooOftenOpenQ = regexp.MustCompile(`^["“„»«]$`)
	styleTooOftenEndQ  = regexp.MustCompile(`^["“”»«]$`)
)

const styleTooOftenMinWordCountDefault = 100

func (r *AbstractStyleTooOftenUsedWordRule) GetID() string {
	if r != nil && r.ID != "" {
		return r.ID
	}
	return "TOO_OFTEN_USED_WORD"
}

func (r *AbstractStyleTooOftenUsedWordRule) GetDescription() string {
	if r != nil && r.Description != "" {
		return r.Description
	}
	return "Word used too often"
}

func (r *AbstractStyleTooOftenUsedWordRule) minWords() int {
	if r == nil {
		return styleTooOftenMinWordCountDefault
	}
	if r.MinWordCount < 0 {
		return styleTooOftenMinWordCountDefault
	}
	return r.MinWordCount
}

// MatchList ports TextLevelRule match: fill word map, then flag tokens over threshold.
func (r *AbstractStyleTooOftenUsedWordRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*RuleMatch {
	if r == nil || r.IsToCountedWord == nil || r.ToAddedLemma == nil {
		return nil
	}
	wordMap := r.fillWordMap(sentences)
	tooOften := r.getTooOftenUsedWords(wordMap)
	if len(tooOften) == 0 {
		return nil
	}
	msgFn := r.LimitMessage
	if msgFn == nil {
		msgFn = func(p int) string { return "Word used too often" }
	}
	msg := msgFn(r.MinPercent)
	var out []*RuleMatch
	pos := 0
	excludeDS := r.WithoutDirectSpeech
	isDirectSpeech := false
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
			if excludeDS && !isDirectSpeech && styleTooOftenOpenQ.MatchString(sToken) &&
				n < len(tokens)-1 && tokens[n+1] != nil && !tokens[n+1].IsWhitespaceBefore() {
				isDirectSpeech = true
			} else if excludeDS && isDirectSpeech && styleTooOftenEndQ.MatchString(sToken) &&
				n > 1 && !token.IsWhitespaceBefore() {
				isDirectSpeech = false
			} else if !isDirectSpeech && !token.IsWhitespace() && !token.IsNonWord() &&
				r.IsToCountedWord(token) && (r.IsException == nil || !r.IsException(tokens, n)) {
				lemma := r.ToAddedLemma(token)
				if lemma == "" {
					continue
				}
				if _, over := tooOften[lemma]; over {
					out = append(out, NewRuleMatch(r, sentence, token.GetStartPos()+pos, token.GetEndPos()+pos, msg))
				}
			}
		}
		pos += sentence.GetCorrectedTextLength()
	}
	return out
}

func (r *AbstractStyleTooOftenUsedWordRule) fillWordMap(sentences []*languagetool.AnalyzedSentence) map[string]int {
	wordMap := map[string]int{}
	excludeDS := r.WithoutDirectSpeech
	isDirectSpeech := false
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
			if excludeDS && !isDirectSpeech && styleTooOftenOpenQ.MatchString(sToken) &&
				n < len(tokens)-1 && tokens[n+1] != nil && !tokens[n+1].IsWhitespaceBefore() {
				isDirectSpeech = true
			} else if excludeDS && isDirectSpeech && styleTooOftenEndQ.MatchString(sToken) &&
				n > 1 && !token.IsWhitespaceBefore() {
				isDirectSpeech = false
			} else if !isDirectSpeech && !token.IsWhitespace() && !token.IsNonWord() &&
				r.IsToCountedWord(token) && (r.IsException == nil || !r.IsException(tokens, n)) {
				lemma := r.ToAddedLemma(token)
				if lemma != "" {
					wordMap[lemma]++
				}
			}
		}
	}
	return wordMap
}

func (r *AbstractStyleTooOftenUsedWordRule) getTooOftenUsedWords(wordMap map[string]int) map[string]struct{} {
	out := map[string]struct{}{}
	numWords := 0
	for _, c := range wordMap {
		numWords += c
	}
	minW := r.minWords()
	// Java: if (numWords < MIN_WORD_COUNT) return empty
	// MinWordCount 0 is only for unit tests that force the gate off (Java constant is 100).
	if minW > 0 && numWords < minW {
		return out
	}
	if numWords == 0 {
		return out
	}
	// Java: percent = (count * 100) / numWords; if (percent >= minPercent) add
	for w, c := range wordMap {
		percent := int(float64(c*100) / float64(numWords))
		if percent >= r.MinPercent {
			out[w] = struct{}{}
		}
	}
	return out
}

// LemmaForPosTagStartsWith ports getLemmaForPosTagStartsWith.
func LemmaForPosTagStartsWith(startPos string, token *languagetool.AnalyzedTokenReadings) string {
	if token == nil {
		return ""
	}
	for _, reading := range token.GetReadings() {
		if reading == nil {
			continue
		}
		pt := reading.GetPOSTag()
		if pt != nil && strings.HasPrefix(*pt, startPos) {
			if l := reading.GetLemma(); l != nil {
				return *l
			}
		}
	}
	return ""
}
