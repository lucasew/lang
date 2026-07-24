package rules

import (
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// SentenceWhitespaceRule ports org.languagetool.rules.SentenceWhitespaceRule.
// Java extends TextLevelRule; constructor sets TYPOGRAPHY + ITSIssueType.Whitespace.
type SentenceWhitespaceRule struct {
	Messages map[string]string
	// RuleID is the Go hook for language-module subclasses that override getId()
	// (e.g. de.SentenceWhitespaceRule → "DE_SENTENCE_WHITESPACE"). Empty → core ID.
	RuleID string
	// Category ports Rule.category (Java Categories.TYPOGRAPHY).
	Category *Category
	// IssueType ports getLocQualityIssueType (Java Whitespace).
	IssueType ITSIssueType
	// MessageAfterSentence / MessageAfterNumber are Go hooks for subclasses that
	// override getMessage(boolean) (de.SentenceWhitespaceRule). Core Java always
	// returns messages.getString("addSpaceBetweenSentences") regardless of the flag.
	MessageAfterSentence string
	MessageAfterNumber   string
	// incorrectExamples / correctExamples port Rule.addExamplePair.
	incorrectExamples []IncorrectExample
	correctExamples   []CorrectExample
}

// NewSentenceWhitespaceRule ports SentenceWhitespaceRule(ResourceBundle messages).
func NewSentenceWhitespaceRule(messages map[string]string) *SentenceWhitespaceRule {
	return &SentenceWhitespaceRule{
		Messages:  messages,
		Category:  CatTypography.GetCategory(messages),
		IssueType: ITSWhitespace,
	}
}

// GetID ports getId() → "SENTENCE_WHITESPACE" (subclasses may set RuleID).
func (r *SentenceWhitespaceRule) GetID() string {
	if r != nil && r.RuleID != "" {
		return r.RuleID
	}
	return "SENTENCE_WHITESPACE"
}

// GetDescription ports getDescription (missing_space_between_sentences).
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

// MinToCheckParagraph ports minToCheckParagraph (Java returns 0).
func (r *SentenceWhitespaceRule) MinToCheckParagraph() int { return 0 }

// EstimateContextForSureMatch ports TextLevelRule (Java always -1).
func (r *SentenceWhitespaceRule) EstimateContextForSureMatch() int { return -1 }

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

// GetMessage ports getMessage(boolean prevSentenceEndsWithNumber).
// Core Java ignores the boolean and returns messages.getString("addSpaceBetweenSentences").
// Subclass hooks (MessageAfter*) map de.SentenceWhitespaceRule overrides.
func (r *SentenceWhitespaceRule) GetMessage(prevEndsWithNumber bool) string {
	if r == nil {
		return "Add a space between sentences"
	}
	if prevEndsWithNumber && r.MessageAfterNumber != "" {
		return r.MessageAfterNumber
	}
	if !prevEndsWithNumber && r.MessageAfterSentence != "" {
		return r.MessageAfterSentence
	}
	if r.Messages != nil {
		if msg := r.Messages["addSpaceBetweenSentences"]; msg != "" {
			return msg
		}
	}
	return "Add a space between sentences"
}

// MatchList ports match(List<AnalyzedSentence>) → RuleMatch[].
func (r *SentenceWhitespaceRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*RuleMatch {
	isFirstSentence := true
	prevSentenceEndsWithWhitespace := false
	prevSentenceEndsWithNumber := false
	var ruleMatches []*RuleMatch
	pos := 0
	for _, sentence := range sentences {
		if sentence == nil {
			continue
		}
		tokens := sentence.GetTokens()
		if isFirstSentence {
			isFirstSentence = false
		} else {
			if !prevSentenceEndsWithWhitespace && len(tokens) > 1 {
				firstToken := tokens[1].GetToken()
				msg := r.GetMessage(prevSentenceEndsWithNumber)
				// Java: new RuleMatch(this, sentence, pos, pos+firstToken.length(), getMessage(...))
				rm := NewRuleMatch(r, sentence, pos, pos+utf16Len(firstToken), msg)
				rm.SetSuggestedReplacement(" " + firstToken)
				ruleMatches = append(ruleMatches, rm)
			}
		}
		if len(tokens) > 0 {
			lastToken := tokens[len(tokens)-1].GetToken()
			// Java: lastToken.replace('\u00A0',' ').trim().isEmpty() && lastToken.length() == 1
			replaced := strings.ReplaceAll(lastToken, "\u00A0", " ")
			prevSentenceEndsWithWhitespace = tools.JavaStringTrimIsEmpty(replaced) && utf16Len(lastToken) == 1
		}
		if len(tokens) > 1 {
			prevLastToken := tokens[len(tokens)-2].GetToken()
			// Java: org.apache.commons.lang3.StringUtils.isNumeric(prevLastToken)
			prevSentenceEndsWithNumber = apacheStringUtilsIsNumeric(prevLastToken)
		}
		pos += sentence.GetCorrectedTextLength()
	}
	return ruleMatches
}

// apacheStringUtilsIsNumeric ports org.apache.commons.lang3.StringUtils.isNumeric:
// non-empty and every char is Character.isDigit (Unicode decimal digits).
func apacheStringUtilsIsNumeric(s string) bool {
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
