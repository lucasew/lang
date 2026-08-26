package rules

import (
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// WhitespaceBeforePunctuationRule ports org.languagetool.rules.WhitespaceBeforePunctuationRule.
// Java: Categories.TYPOGRAPHY, ITSIssueType.Whitespace.
type WhitespaceBeforePunctuationRule struct {
	Messages  map[string]string
	Category  *Category
	IssueType ITSIssueType
}

// NewWhitespaceBeforePunctuationRule ports WhitespaceBeforePunctuationRule(ResourceBundle messages).
func NewWhitespaceBeforePunctuationRule(messages map[string]string) *WhitespaceBeforePunctuationRule {
	return &WhitespaceBeforePunctuationRule{
		Messages:  messages,
		Category:  CatTypography.GetCategory(messages),
		IssueType: ITSWhitespace,
	}
}

// GetID ports getId() → "WHITESPACE_PUNCTUATION".
func (r *WhitespaceBeforePunctuationRule) GetID() string { return "WHITESPACE_PUNCTUATION" }

// GetDescription ports getDescription (desc_whitespace_before_punctuation).
func (r *WhitespaceBeforePunctuationRule) GetDescription() string {
	if r != nil && r.Messages != nil {
		if s := r.Messages["desc_whitespace_before_punctuation"]; s != "" {
			return s
		}
	}
	// MessagesBundle_en.properties
	return "Use of whitespace before colon, semicolon and percentage."
}

func (r *WhitespaceBeforePunctuationRule) GetCategory() *Category {
	if r == nil {
		return nil
	}
	return r.Category
}

func (r *WhitespaceBeforePunctuationRule) GetLocQualityIssueType() ITSIssueType {
	if r == nil || r.IssueType == "" {
		return ITSWhitespace
	}
	return r.IssueType
}

// Match ports WhitespaceBeforePunctuationRule.match(AnalyzedSentence) → RuleMatch[].
// Control flow is a line-by-line transcription of the Java method (bug-for-bug),
// including toPos = startPos + 1 + prevLen and the " : 0" colon exception.
func (r *WhitespaceBeforePunctuationRule) Match(sentence *languagetool.AnalyzedSentence) []*RuleMatch {
	var ruleMatches []*RuleMatch
	tokens := sentence.GetTokens()
	prevWhite := false
	prevLen := 0
	for i := 0; i < len(tokens); i++ {
		token := tokens[i].GetToken()
		// Java: isWhitespace || StringTools.isNonBreakingWhitespace(token) || isFieldCode()
		isWhitespace := tokens[i].IsWhitespace() || tools.IsNonBreakingWhitespace(token) || tokens[i].IsFieldCode()
		// msg empty == Java null
		msg := ""
		suggestionText := ""
		if prevWhite {
			if token == ":" {
				msg = r.msg("no_space_before_colon", "Don't put a space before the colon.")
				suggestionText = ":"
				// exception case for figures such as " : 0"
				if i+2 < len(tokens) &&
					tokens[i+1].IsWhitespace() &&
					javaCharIsDigitAt0(tokens[i+2].GetToken()) {
					msg = ""
				}
			} else if token == ";" {
				msg = r.msg("no_space_before_semicolon", "Don't put a space before the semicolon.")
				suggestionText = ";"
			} else if i > 1 && token == "%" {
				prevPrevToken := tokens[i-2].GetToken()
				// Java: prevPrevToken.length() > 0 && Character.isDigit(prevPrevToken.charAt(0))
				if utf16Len(prevPrevToken) > 0 && javaCharIsDigitAt0(prevPrevToken) {
					msg = r.msg("no_space_before_percentage", "Don't put a space before the percentage.")
					suggestionText = "%"
				}
			}
		}
		if msg != "" {
			// Java: fromPos = tokens[i-1].getStartPos(); toPos = fromPos + 1 + prevLen
			fromPos := tokens[i-1].GetStartPos()
			toPos := tokens[i-1].GetStartPos() + 1 + prevLen
			ruleMatch := NewRuleMatch(r, sentence, fromPos, toPos, msg)
			ruleMatch.SetSuggestedReplacement(suggestionText)
			ruleMatches = append(ruleMatches, ruleMatch)
		}
		// Java: prevWhite = isWhitespace && !tokens[i].isFieldCode(); //OOo code before comma/dot
		prevWhite = isWhitespace && !tokens[i].IsFieldCode()
		// Java: prevLen = tokens[i].getToken().length() (UTF-16 code units)
		prevLen = utf16Len(tokens[i].GetToken())
	}
	return ruleMatches
}

func (r *WhitespaceBeforePunctuationRule) msg(key, def string) string {
	if r != nil && r.Messages != nil {
		if s := r.Messages[key]; s != "" {
			return s
		}
	}
	return def
}

// javaCharIsDigitAt0 ports Character.isDigit(s.charAt(0)).
// Empty string returns false (Java would throw IndexOutOfBoundsException).
// Non-BMP first code point: Java charAt(0) is a high surrogate → isDigit false.
func javaCharIsDigitAt0(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r >= 0x10000 {
			return false
		}
		return unicode.IsDigit(r)
	}
	return false
}
