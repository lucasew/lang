package rules

import (
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// SentenceWhitespaceRule ports org.languagetool.rules.SentenceWhitespaceRule.
type SentenceWhitespaceRule struct {
	Messages map[string]string
}

func NewSentenceWhitespaceRule(messages map[string]string) *SentenceWhitespaceRule {
	return &SentenceWhitespaceRule{Messages: messages}
}

func (r *SentenceWhitespaceRule) GetID() string { return "SENTENCE_WHITESPACE" }

func (r *SentenceWhitespaceRule) GetMessage(prevEndsWithNumber bool) string {
	msg := r.Messages["addSpaceBetweenSentences"]
	if msg == "" {
		msg = "Add a space between sentences"
	}
	return msg
}

// MatchList ports match(List<AnalyzedSentence>).
func (r *SentenceWhitespaceRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*RuleMatch {
	isFirstSentence := true
	prevSentenceEndsWithWhitespace := false
	prevSentenceEndsWithNumber := false
	var ruleMatches []*RuleMatch
	pos := 0
	for _, sentence := range sentences {
		tokens := sentence.GetTokens()
		if isFirstSentence {
			isFirstSentence = false
		} else {
			if !prevSentenceEndsWithWhitespace && len(tokens) > 1 {
				firstToken := tokens[1].GetToken()
				msg := r.GetMessage(prevSentenceEndsWithNumber)
				rm := NewRuleMatch(r, sentence, pos, pos+utf16Len(firstToken), msg)
				rm.SetSuggestedReplacement(" " + firstToken)
				ruleMatches = append(ruleMatches, rm)
			}
		}
		if len(tokens) > 0 {
			lastToken := tokens[len(tokens)-1].GetToken()
			replaced := strings.ReplaceAll(lastToken, "\u00A0", " ")
			prevSentenceEndsWithWhitespace = strings.TrimSpace(replaced) == "" && len([]rune(lastToken)) == 1
			// Java: lastToken.length() == 1 (UTF-16)
			prevSentenceEndsWithWhitespace = strings.TrimSpace(replaced) == "" && utf16Len(lastToken) == 1
		}
		if len(tokens) > 1 {
			prevLastToken := tokens[len(tokens)-2].GetToken()
			prevSentenceEndsWithNumber = isNumeric(prevLastToken)
		}
		pos += sentence.GetCorrectedTextLength()
	}
	return ruleMatches
}

func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}
