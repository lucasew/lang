package rules

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// MultipleWhitespaceRule ports org.languagetool.rules.MultipleWhitespaceRule.
// MultipleWhitespaceRule ports org.languagetool.rules.MultipleWhitespaceRule.
// Java: TYPOGRAPHY, Whitespace.
type MultipleWhitespaceRule struct {
	Messages  map[string]string
	Category  *Category
	IssueType ITSIssueType
}

func NewMultipleWhitespaceRule(messages map[string]string) *MultipleWhitespaceRule {
	return &MultipleWhitespaceRule{
		Messages:  messages,
		Category:  CatTypography.GetCategory(messages),
		IssueType: ITSWhitespace,
	}
}

func (r *MultipleWhitespaceRule) GetID() string { return "WHITESPACE_RULE" }

// GetDescription ports getDescription (desc_whitespacerepetition).
func (r *MultipleWhitespaceRule) GetDescription() string {
	if r != nil && r.Messages != nil {
		if s := r.Messages["desc_whitespacerepetition"]; s != "" {
			return s
		}
	}
	return "Whitespace repetition (bad formatting)"
}

func (r *MultipleWhitespaceRule) GetCategory() *Category {
	if r == nil {
		return nil
	}
	return r.Category
}

func (r *MultipleWhitespaceRule) GetLocQualityIssueType() ITSIssueType {
	if r == nil || r.IssueType == "" {
		return ITSWhitespace
	}
	return r.IssueType
}

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
		if len(sub) > 0 && containsSubstr(s, sub) {
			return true
		}
	}
	return false
}

func containsSubstr(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || indexOfByte(s, sub) >= 0)
}

func indexOfByte(s, sub string) int {
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
