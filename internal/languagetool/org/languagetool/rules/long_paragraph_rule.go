package rules

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

var nonWordRE = regexp.MustCompile(`^[.?!…:;,~’'"„“”»«‚‘›‹()\[\]\-–—*×∗·+÷/=]$`)

// LongParagraphRule ports org.languagetool.rules.LongParagraphRule.
type LongParagraphRule struct {
	Messages map[string]string
	MaxWords int
	// SingleLineBreaksMarksPara matches Demo/SRX default false → need \n\n
	SingleLineBreaksMarksPara bool
}

func NewLongParagraphRule(messages map[string]string, maxWords int) *LongParagraphRule {
	return &LongParagraphRule{Messages: messages, MaxWords: maxWords}
}

func (r *LongParagraphRule) GetID() string { return "TOO_LONG_PARAGRAPH" }

func (r *LongParagraphRule) GetMessage() string {
	msg := r.Messages["long_paragraph_rule_msg"]
	if msg == "" {
		msg = "This paragraph is too long (%d words)"
	}
	if strings.Contains(msg, "%d") || strings.Contains(msg, "{0}") {
		return fmt.Sprintf(strings.ReplaceAll(msg, "{0}", "%d"), r.MaxWords)
	}
	return fmt.Sprintf(msg, r.MaxWords)
}

func (r *LongParagraphRule) isParagraphEnd(sentences []*languagetool.AnalyzedSentence, nTest int) bool {
	if nTest >= len(sentences)-1 {
		return true
	}
	text := sentences[nTest].GetText()
	if r.SingleLineBreaksMarksPara {
		if strings.HasSuffix(text, "\n") || strings.HasSuffix(text, "\n\r") {
			return true
		}
	} else {
		if strings.HasSuffix(text, "\n\n") || strings.HasSuffix(text, "\n\r\n\r") || strings.HasSuffix(text, "\r\n\r\n") {
			return true
		}
	}
	next := sentences[nTest+1].GetText()
	if strings.HasPrefix(next, "\n") || strings.HasPrefix(next, "\r\n") {
		return true
	}
	return false
}

// MatchList ports match(List<AnalyzedSentence>).
func (r *LongParagraphRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*RuleMatch {
	var ruleMatches []*RuleMatch
	pos := 0
	startPos, endPos := 0, 0
	wordCount := 0
	paraHasLinebreaks := false
	for n := 0; n < len(sentences); n++ {
		sentence := sentences[n]
		paragraphEnd := r.isParagraphEnd(sentences, n)
		trimmed := regexp.MustCompile(`^\n+`).ReplaceAllString(sentence.GetText(), "")
		if !paragraphEnd && strings.Contains(trimmed, "\n") {
			paraHasLinebreaks = true
		}
		tokens := sentence.GetTokensWithoutWhitespace()
		for _, token := range tokens {
			if !token.IsWhitespace() && !token.IsSentenceStart() && !isNonWordToken(token) {
				wordCount++
				if wordCount == r.MaxWords {
					endPos = token.GetEndPos() + pos
				} else if wordCount == r.MaxWords-1 {
					startPos = token.GetStartPos() + pos
				}
			}
		}
		if paragraphEnd {
			if wordCount > r.MaxWords+5 && !paraHasLinebreaks {
				// last sentence of paragraph used for match attachment
				rm := NewRuleMatch(r, sentence, startPos, endPos, r.GetMessage())
				ruleMatches = append(ruleMatches, rm)
			}
			wordCount = 0
			paraHasLinebreaks = false
		}
		pos += sentence.GetCorrectedTextLength()
	}
	if wordCount > r.MaxWords {
		// no sentence ref in Java constructor without sentence - use last
		var last *languagetool.AnalyzedSentence
		if len(sentences) > 0 {
			last = sentences[len(sentences)-1]
		}
		rm := NewRuleMatch(r, last, startPos, endPos, r.GetMessage())
		ruleMatches = append(ruleMatches, rm)
	}
	return ruleMatches
}

func isNonWordToken(token *languagetool.AnalyzedTokenReadings) bool {
	return nonWordRE.MatchString(token.GetToken())
}
