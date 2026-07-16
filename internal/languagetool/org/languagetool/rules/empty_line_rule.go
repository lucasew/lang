package rules

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// EmptyLineRule ports org.languagetool.rules.EmptyLineRule.
type EmptyLineRule struct {
	Messages                  map[string]string
	SingleLineBreaksMarksPara bool
}

func NewEmptyLineRule(messages map[string]string) *EmptyLineRule {
	return &EmptyLineRule{Messages: messages}
}

func (r *EmptyLineRule) GetID() string { return "EMPTY_LINE" }

func (r *EmptyLineRule) isParagraphEnd(sentences []*languagetool.AnalyzedSentence, nTest int) bool {
	if nTest >= len(sentences)-1 {
		return true
	}
	text := sentences[nTest].GetText()
	if r.SingleLineBreaksMarksPara {
		if len(text) > 0 && text[len(text)-1] == '\n' {
			return true
		}
	} else {
		if hasSuf(text, "\n\n") || hasSuf(text, "\n\r\n\r") || hasSuf(text, "\r\n\r\n") {
			return true
		}
	}
	next := sentences[nTest+1].GetText()
	if len(next) > 0 && (next[0] == '\n' || hasPre(next, "\r\n")) {
		return true
	}
	return false
}

func hasSuf(s, suf string) bool {
	return len(s) >= len(suf) && s[len(s)-len(suf):] == suf
}
func hasPre(s, pre string) bool {
	return len(s) >= len(pre) && s[:len(pre)] == pre
}

func (r *EmptyLineRule) isSecondParagraphEndMark(sentence string) bool {
	if r.SingleLineBreaksMarksPara {
		return hasSuf(sentence, "\n\n") || hasSuf(sentence, "\n\r\n\r")
	}
	return hasSuf(sentence, "\n\n\n\n") || hasSuf(sentence, "\n\r\n\r\n\r\n\r") || hasSuf(sentence, "\r\n\r\n\r\n\r\n")
}

// MatchList ports match(List<AnalyzedSentence>).
func (r *EmptyLineRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*RuleMatch {
	var ruleMatches []*RuleMatch
	pos := 0
	msg := r.Messages["empty_line_rule_msg"]
	if msg == "" {
		msg = "Empty line"
	}
	for n := 0; n < len(sentences)-1; n++ {
		sentence := sentences[n]
		if r.isParagraphEnd(sentences, n) && r.isSecondParagraphEndMark(sentence.GetText()) {
			tokens := sentence.GetTokensWithoutWhitespace()
			if len(tokens) > 1 {
				fromPos := pos + tokens[len(tokens)-1].GetStartPos()
				toPos := pos + tokens[len(tokens)-1].GetEndPos()
				ruleMatches = append(ruleMatches, NewRuleMatch(r, sentence, fromPos, toPos, msg))
			}
		}
		pos += sentence.GetCorrectedTextLength()
	}
	return ruleMatches
}
