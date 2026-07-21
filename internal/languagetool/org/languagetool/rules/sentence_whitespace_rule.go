package rules

import (
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// SentenceWhitespaceRule ports org.languagetool.rules.SentenceWhitespaceRule.
// Java: TYPOGRAPHY, Whitespace.
type SentenceWhitespaceRule struct {
	Messages map[string]string
	// RuleID overrides GetID when set (e.g. DE_SENTENCE_WHITESPACE).
	RuleID string
	// Category ports Rule.category (Java TYPOGRAPHY).
	Category *Category
	// IssueType ports getLocQualityIssueType (Java Whitespace).
	IssueType ITSIssueType
	// MessageAfterSentence / MessageAfterNumber override default messages.
	MessageAfterSentence string
	MessageAfterNumber   string
	// incorrectExamples / correctExamples port Rule.addExamplePair.
	incorrectExamples []IncorrectExample
	correctExamples   []CorrectExample
}

func NewSentenceWhitespaceRule(messages map[string]string) *SentenceWhitespaceRule {
	return &SentenceWhitespaceRule{
		Messages:  messages,
		Category:  CatTypography.GetCategory(messages),
		IssueType: ITSWhitespace,
	}
}

func (r *SentenceWhitespaceRule) GetID() string {
	if r.RuleID != "" {
		return r.RuleID
	}
	return "SENTENCE_WHITESPACE"
}

// GetDescription ports core getDescription (missing_space_between_sentences).
// DE overrides with language-specific text.
func (r *SentenceWhitespaceRule) GetDescription() string {
	if r != nil && r.Messages != nil {
		if s := r.Messages["missing_space_between_sentences"]; s != "" {
			return s
		}
	}
	return "Missing space between sentences"
}

func (r *SentenceWhitespaceRule) GetCategory() *Category {
	if r == nil {
		return nil
	}
	return r.Category
}

func (r *SentenceWhitespaceRule) GetLocQualityIssueType() ITSIssueType {
	if r == nil || r.IssueType == "" {
		return ITSWhitespace
	}
	return r.IssueType
}

// AddExamplePair ports Rule.addExamplePair.
func (r *SentenceWhitespaceRule) AddExamplePair(incorrect IncorrectExample, correct CorrectExample) {
	if r == nil {
		return
	}
	appendExamplePair(&r.incorrectExamples, &r.correctExamples, incorrect, correct)
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *SentenceWhitespaceRule) GetIncorrectExamples() []IncorrectExample {
	if r == nil || len(r.incorrectExamples) == 0 {
		return nil
	}
	out := make([]IncorrectExample, len(r.incorrectExamples))
	copy(out, r.incorrectExamples)
	return out
}

// GetCorrectExamples ports Rule.getCorrectExamples.
func (r *SentenceWhitespaceRule) GetCorrectExamples() []CorrectExample {
	if r == nil || len(r.correctExamples) == 0 {
		return nil
	}
	out := make([]CorrectExample, len(r.correctExamples))
	copy(out, r.correctExamples)
	return out
}

func (r *SentenceWhitespaceRule) GetMessage(prevEndsWithNumber bool) string {
	if prevEndsWithNumber && r.MessageAfterNumber != "" {
		return r.MessageAfterNumber
	}
	if !prevEndsWithNumber && r.MessageAfterSentence != "" {
		return r.MessageAfterSentence
	}
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
			// Java: lastToken.replace('\u00A0',' ').trim().isEmpty() && lastToken.length() == 1
			prevSentenceEndsWithWhitespace = tools.JavaStringTrimIsEmpty(replaced) && utf16Len(lastToken) == 1
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
