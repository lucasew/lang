package rules

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// MultipleWhitespaceRule ports org.languagetool.rules.MultipleWhitespaceRule.
// Java: TYPOGRAPHY, Whitespace.
type MultipleWhitespaceRule struct {
	Messages  map[string]string
	Category  *Category
	IssueType ITSIssueType
}

// NewMultipleWhitespaceRule ports MultipleWhitespaceRule(ResourceBundle, Language).
// Language is unused in Java match; messages drive description and match text.
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

// MinToCheckParagraph ports MultipleWhitespaceRule.minToCheckParagraph (Java returns 0).
func (r *MultipleWhitespaceRule) MinToCheckParagraph() int { return 0 }

// isFirstWhite ports MultipleWhitespaceRule.isFirstWhite:
// whitespace or NBSP, not linebreak, and not containing ZWSP/BOM/WJ.
func isFirstWhite(token *languagetool.AnalyzedTokenReadings) bool {
	t := token.GetToken()
	return (token.IsWhitespace() || tools.IsNonBreakingWhitespace(t)) &&
		!token.IsLinebreak() &&
		!strings.Contains(t, "\u200B") && !strings.Contains(t, "\uFEFF") && !strings.Contains(t, "\u2060")
}

// isRemovableWhite ports MultipleWhitespaceRule.isRemovableWhite:
// same as isFirstWhite but also excludes tabs.
func isRemovableWhite(token *languagetool.AnalyzedTokenReadings) bool {
	t := token.GetToken()
	return (token.IsWhitespace() || tools.IsNonBreakingWhitespace(t)) &&
		!token.IsLinebreak() && t != "\t" &&
		!strings.Contains(t, "\u200B") && !strings.Contains(t, "\uFEFF") && !strings.Contains(t, "\u2060")
}

// Match ports TextLevelRule match(List<AnalyzedSentence>).
// Control flow mirrors Java including the linebreak branch that advances i without i--.
func (r *MultipleWhitespaceRule) Match(sentences []*languagetool.AnalyzedSentence) []*RuleMatch {
	var ruleMatches []*RuleMatch
	pos := 0
	msg := ""
	if r != nil && r.Messages != nil {
		msg = r.Messages["whitespace_repetition"]
	}
	if msg == "" {
		msg = "Possible typo: you repeated a whitespace"
	}
	for _, sentence := range sentences {
		tokens := sentence.GetTokens()
		// note: we start from token 1
		// token no. 0 is guaranteed to be SENT_START
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
				// Java: for (i++; …; i++); — no i--, so outer loop advances past next token.
				for i++; i < len(tokens) && isRemovableWhite(tokens[i]); i++ {
				}
			}
		}
		pos += sentence.GetCorrectedTextLength()
	}
	return ruleMatches
}
