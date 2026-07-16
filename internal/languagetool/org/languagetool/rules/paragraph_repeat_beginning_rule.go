package rules

import (
	"regexp"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// ParagraphRepeatBeginningRule ports org.languagetool.rules.ParagraphRepeatBeginningRule.
type ParagraphRepeatBeginningRule struct {
	Messages map[string]string
	// IsArticle optional POS stand-in; default false (only first word compared).
	IsArticle func(token *languagetool.AnalyzedTokenReadings) bool
	// SingleLineBreaksMarksPara matches Demo/SRX default false → need \n\n
	SingleLineBreaksMarksPara bool
	RuleID                    string
}

var paraQuotesRE = regexp.MustCompile(`[’'"„“”»«‚‘›‹()\[\]]`)

func NewParagraphRepeatBeginningRule(messages map[string]string) *ParagraphRepeatBeginningRule {
	return &ParagraphRepeatBeginningRule{Messages: messages}
}

func (r *ParagraphRepeatBeginningRule) GetID() string {
	if r.RuleID != "" {
		return r.RuleID
	}
	return "PARAGRAPH_REPEAT_BEGINNING_RULE"
}

func (r *ParagraphRepeatBeginningRule) isParagraphEnd(sentences []*languagetool.AnalyzedSentence, nTest int) bool {
	if nTest >= len(sentences)-1 {
		return true
	}
	text := sentences[nTest].GetText()
	if r.SingleLineBreaksMarksPara {
		if len(text) > 0 && (text[len(text)-1] == '\n' || hasSuffix(text, "\n\r")) {
			return true
		}
	} else {
		if hasSuffix(text, "\n\n") || hasSuffix(text, "\n\r\n\r") || hasSuffix(text, "\r\n\r\n") {
			return true
		}
	}
	next := sentences[nTest+1].GetText()
	if len(next) > 0 && (next[0] == '\n' || hasPrefix(next, "\r\n")) {
		return true
	}
	return false
}

func hasSuffix(s, suf string) bool {
	return len(s) >= len(suf) && s[len(s)-len(suf):] == suf
}
func hasPrefix(s, pre string) bool {
	return len(s) >= len(pre) && s[:len(pre)] == pre
}

func (r *ParagraphRepeatBeginningRule) isArticle(token *languagetool.AnalyzedTokenReadings) bool {
	if r.IsArticle != nil {
		return r.IsArticle(token)
	}
	return false
}

// numCharEqualBeginning returns endPos of matching beginning token in lastTokens, or 0.
func (r *ParagraphRepeatBeginningRule) numCharEqualBeginning(lastTokens, nextTokens []*languagetool.AnalyzedTokenReadings) int {
	if len(lastTokens) < 2 || len(nextTokens) < 2 {
		return 0
	}
	nToken := 1
	lastToken := lastTokens[nToken].GetToken()
	nextToken := nextTokens[nToken].GetToken()
	if paraQuotesRE.MatchString(lastToken) && lastToken == nextToken {
		if len(lastTokens) <= nToken+1 || len(nextTokens) <= nToken+1 {
			return 0
		}
		nToken++
		lastToken = lastTokens[nToken].GetToken()
		nextToken = nextTokens[nToken].GetToken()
	}
	if lastToken == "" || !unicode.IsLetter([]rune(lastToken)[0]) {
		return 0
	}
	if len(lastTokens) > nToken+1 && r.isArticle(lastTokens[nToken]) && lastToken == nextToken {
		if len(nextTokens) <= nToken+1 {
			return 0
		}
		nToken++
		lastToken = lastTokens[nToken].GetToken()
		nextToken = nextTokens[nToken].GetToken()
	}
	if lastToken == "" || !unicode.IsLetter([]rune(lastToken)[0]) {
		return 0
	}
	if lastToken == nextToken {
		return lastTokens[nToken].GetEndPos()
	}
	return 0
}

// MatchList ports match(List<AnalyzedSentence>).
func (r *ParagraphRepeatBeginningRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*RuleMatch {
	var ruleMatches []*RuleMatch
	if len(sentences) < 1 {
		return ruleMatches
	}
	nextPos := 0
	lastPos := 0
	lastSentence := sentences[0]
	lastTokens := lastSentence.GetTokensWithoutWhitespace()
	msg := r.Messages["repetition_paragraph_beginning_last_msg"]
	if msg == "" {
		msg = "Paragraphs should not begin with the same words"
	}
	for n := 0; n < len(sentences)-1; n++ {
		// Java uses getText().length() (UTF-16) — use CorrectedTextLength for consistency with other ports
		nextPos += sentences[n].GetCorrectedTextLength()
		if !r.isParagraphEnd(sentences, n) {
			continue
		}
		nextSentence := sentences[n+1]
		nextTokens := nextSentence.GetTokensWithoutWhitespace()
		endPos := r.numCharEqualBeginning(lastTokens, nextTokens)
		if endPos > 0 {
			startPos := lastPos + lastTokens[1].GetStartPos()
			if startPos < lastPos+endPos {
				rm := NewRuleMatch(r, lastSentence, startPos, lastPos+endPos, msg)
				ruleMatches = append(ruleMatches, rm)
				startPos2 := nextPos + nextTokens[1].GetStartPos()
				rm2 := NewRuleMatch(r, nextSentence, startPos2, nextPos+endPos, msg)
				ruleMatches = append(ruleMatches, rm2)
			}
		}
		lastSentence = nextSentence
		lastTokens = nextTokens
		lastPos = nextPos
	}
	return ruleMatches
}
