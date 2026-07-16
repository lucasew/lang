package rules

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// MultipleWhitespaceRule ports org.languagetool.rules.MultipleWhitespaceRule.
type MultipleWhitespaceRule struct {
	Messages map[string]string
}

func NewMultipleWhitespaceRule(messages map[string]string) *MultipleWhitespaceRule {
	return &MultipleWhitespaceRule{Messages: messages}
}

func (r *MultipleWhitespaceRule) GetID() string { return "WHITESPACE_RULE" }

func isFirstWhite(token *languagetool.AnalyzedTokenReadings) bool {
	t := token.GetToken()
	return (token.IsWhitespace() || tools.IsNonBreakingWhitespace(t)) &&
		!token.IsLinebreak() &&
		!containsAny(t, "\u200B", "\uFEFF", "\u2060")
}

func isRemovableWhite(token *languagetool.AnalyzedTokenReadings) bool {
	t := token.GetToken()
	return (token.IsWhitespace() || tools.IsNonBreakingWhitespace(t)) &&
		!token.IsLinebreak() && t != "\t" &&
		!containsAny(t, "\u200B", "\uFEFF", "\u2060")
}

func containsAny(s string, subs ...string) bool {
	for _, sub := range subs {
		if len(sub) > 0 && containsStr(s, sub) {
			return true
		}
	}
	return false
}

func containsStr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

// Match ports TextLevelRule match over sentences.
func (r *MultipleWhitespaceRule) Match(sentences []*languagetool.AnalyzedSentence) []*RuleMatch {
	var ruleMatches []*RuleMatch
	pos := 0
	msg := r.Messages["whitespace_repetition"]
	if msg == "" {
		msg = "Possible typo: you repeated a whitespace"
	}
	for _, sentence := range sentences {
		tokens := sentence.GetTokens()
		for i := 1; i < len(tokens); i++ {
			if isFirstWhite(tokens[i]) {
				nFirst := i
				for i++; i < len(tokens) && isRemovableWhite(tokens[i]); i++ {
				}
				i--
				if i > nFirst {
					from := pos + tokens[nFirst].GetStartPos()
					to := pos + tokens[i].GetEndPos()
					rm := NewRuleMatch(r, sentence, from, to, msg)
					rm.SetSuggestedReplacement(tokens[nFirst].GetToken())
					ruleMatches = append(ruleMatches, rm)
				}
			} else if tokens[i].IsLinebreak() {
				for i++; i < len(tokens) && isRemovableWhite(tokens[i]); i++ {
				}
				i--
			}
		}
		pos += sentence.GetCorrectedTextLength()
	}
	return ruleMatches
}
