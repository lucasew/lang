package rules

import (
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// WhitespaceBeforePunctuationRule ports org.languagetool.rules.WhitespaceBeforePunctuationRule.
type WhitespaceBeforePunctuationRule struct {
	Messages map[string]string
}

func NewWhitespaceBeforePunctuationRule(messages map[string]string) *WhitespaceBeforePunctuationRule {
	return &WhitespaceBeforePunctuationRule{Messages: messages}
}

func (r *WhitespaceBeforePunctuationRule) GetID() string { return "WHITESPACE_PUNCTUATION" }

func (r *WhitespaceBeforePunctuationRule) Match(sentence *languagetool.AnalyzedSentence) []*RuleMatch {
	var ruleMatches []*RuleMatch
	tokens := sentence.GetTokens()
	prevWhite := false
	prevLen := 0
	for i := 0; i < len(tokens); i++ {
		token := tokens[i].GetToken()
		isWhitespace := tokens[i].IsWhitespace() || tools.IsNonBreakingWhitespace(token) || tokens[i].IsFieldCode()
		var msg, suggestionText string
		msgSet := false
		if prevWhite {
			if token == ":" {
				msg = r.msg("no_space_before_colon", "Don't put a space before the colon")
				suggestionText = ":"
				msgSet = true
				if i+2 < len(tokens) && tokens[i+1].IsWhitespace() {
					next := tokens[i+2].GetToken()
					if next != "" {
						r0 := []rune(next)[0]
						if unicode.IsDigit(r0) {
							msgSet = false
						}
					}
				}
			} else if token == ";" {
				msg = r.msg("no_space_before_semicolon", "Don't put a space before the semicolon")
				suggestionText = ";"
				msgSet = true
			} else if i > 1 && token == "%" {
				prevPrevToken := tokens[i-2].GetToken()
				if prevPrevToken != "" {
					r0 := []rune(prevPrevToken)[0]
					if unicode.IsDigit(r0) {
						msg = r.msg("no_space_before_percentage", "Don't put a space before the percentage sign")
						suggestionText = "%"
						msgSet = true
					}
				}
			}
		}
		if msgSet {
			fromPos := tokens[i-1].GetStartPos()
			toPos := tokens[i-1].GetStartPos() + 1 + prevLen
			rm := NewRuleMatch(r, sentence, fromPos, toPos, msg)
			rm.SetSuggestedReplacement(suggestionText)
			ruleMatches = append(ruleMatches, rm)
		}
		prevWhite = isWhitespace && !tokens[i].IsFieldCode()
		// Java: prevLen = tokens[i].getToken().length() UTF-16
		prevLen = 0
		for _, r := range tokens[i].GetToken() {
			if r >= 0x10000 {
				prevLen += 2
			} else {
				prevLen++
			}
		}
	}
	return ruleMatches
}

func (r *WhitespaceBeforePunctuationRule) msg(key, def string) string {
	if r.Messages != nil {
		if s, ok := r.Messages[key]; ok && s != "" {
			return s
		}
	}
	return def
}
