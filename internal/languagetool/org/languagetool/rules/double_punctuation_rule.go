package rules

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// DoublePunctuationRule ports org.languagetool.rules.DoublePunctuationRule.
// Java: PUNCTUATION, Typographical.
type DoublePunctuationRule struct {
	Messages       map[string]string
	RuleID         string // override GetID when set (e.g. DE_DOUBLE_PUNCTUATION)
	DotMessage     string // override two-dots message when set
	CommaCharacter string // override comma character (Arabic/Persian "،")
	Category       *Category
	IssueType      ITSIssueType
	// URL ports Rule.url (Java setUrl).
	URL string
	// incorrectExamples / correctExamples port Rule.addExamplePair.
	incorrectExamples []IncorrectExample
	correctExamples   []CorrectExample
}

func NewDoublePunctuationRule(messages map[string]string) *DoublePunctuationRule {
	return &DoublePunctuationRule{
		Messages:  messages,
		Category:  CatPunctuation.GetCategory(messages),
		IssueType: ITSTypographical,
	}
}

func (r *DoublePunctuationRule) GetCategory() *Category {
	if r == nil {
		return nil
	}
	return r.Category
}

func (r *DoublePunctuationRule) GetLocQualityIssueType() ITSIssueType {
	if r == nil || r.IssueType == "" {
		return ITSTypographical
	}
	return r.IssueType
}

// GetURL ports Rule.getUrl.
func (r *DoublePunctuationRule) GetURL() string {
	if r == nil {
		return ""
	}
	return r.URL
}

// AddExamplePair ports Rule.addExamplePair.
func (r *DoublePunctuationRule) AddExamplePair(incorrect IncorrectExample, correct CorrectExample) {
	if r == nil {
		return
	}
	appendExamplePair(&r.incorrectExamples, &r.correctExamples, incorrect, correct)
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *DoublePunctuationRule) GetIncorrectExamples() []IncorrectExample {
	if r == nil || len(r.incorrectExamples) == 0 {
		return nil
	}
	out := make([]IncorrectExample, len(r.incorrectExamples))
	copy(out, r.incorrectExamples)
	return out
}

// GetCorrectExamples ports Rule.getCorrectExamples.
func (r *DoublePunctuationRule) GetCorrectExamples() []CorrectExample {
	if r == nil || len(r.correctExamples) == 0 {
		return nil
	}
	out := make([]CorrectExample, len(r.correctExamples))
	copy(out, r.correctExamples)
	return out
}

func (r *DoublePunctuationRule) GetID() string {
	if r.RuleID != "" {
		return r.RuleID
	}
	return "DOUBLE_PUNCTUATION"
}

// GetDescription ports DoublePunctuationRule.getDescription (desc_double_punct).
func (r *DoublePunctuationRule) GetDescription() string {
	if r != nil && r.Messages != nil {
		if s := r.Messages["desc_double_punct"]; s != "" {
			return s
		}
	}
	return "Two consecutive dots or commas"
}

func (r *DoublePunctuationRule) GetCommaCharacter() string {
	if r.CommaCharacter != "" {
		return r.CommaCharacter
	}
	return ","
}

// getDotMessage ports DoublePunctuationRule.getDotMessage.
func (r *DoublePunctuationRule) getDotMessage() string {
	if r.DotMessage != "" {
		return r.DotMessage
	}
	if r.Messages != nil {
		if s := r.Messages["two_dots"]; s != "" {
			return s
		}
	}
	return "Two consecutive dots"
}

// getCommaMessage ports DoublePunctuationRule.getCommaMessage.
func (r *DoublePunctuationRule) getCommaMessage() string {
	if r.Messages != nil {
		if s := r.Messages["two_commas"]; s != "" {
			return s
		}
	}
	return "Two consecutive commas"
}

func (r *DoublePunctuationRule) msg(key, def string) string {
	if r != nil && r.Messages != nil {
		if s := r.Messages[key]; s != "" {
			return s
		}
	}
	return def
}

func (r *DoublePunctuationRule) Match(sentence *languagetool.AnalyzedSentence) []*RuleMatch {
	var ruleMatches []*RuleMatch
	tokens := sentence.GetTokensWithoutWhitespace()
	startPos := 0
	dotCount, commaCount := 0, 0
	commaChar := r.GetCommaCharacter()
	for i := 1; i < len(tokens); i++ {
		token := tokens[i].GetToken()
		var nextToken, prevPrevToken string
		if i < len(tokens)-1 {
			nextToken = tokens[i+1].GetToken()
		}
		if i > 1 {
			prevPrevToken = tokens[i-2].GetToken()
		}
		if token == "." {
			dotCount++
			commaCount = 0
			startPos = tokens[i].GetStartPos()
		} else if token == commaChar {
			commaCount++
			dotCount = 0
			startPos = tokens[i].GetStartPos()
		}

		if dotCount == 2 && nextToken != "." && nextToken != "…" &&
			token != "/" && nextToken != "/" &&
			token != "\\" && nextToken != "\\" &&
			prevPrevToken != "?" && prevPrevToken != "!" &&
			prevPrevToken != "…" && prevPrevToken != "." {
			fromPos := startPos - 1
			if fromPos < 0 {
				fromPos = 0
			}
			// Java: new RuleMatch(..., getDotMessage(), messages.getString("double_dots_short"))
			rm := NewRuleMatch(r, sentence, fromPos, startPos+1, r.getDotMessage())
			rm.ShortMessage = r.msg("double_dots_short", "")
			rm.SuggestedReplacements = []string{".", "…"}
			ruleMatches = append(ruleMatches, rm)
			dotCount = 0
		} else if commaCount == 2 && nextToken != commaChar {
			fromPos := startPos - 1
			if fromPos < 0 {
				fromPos = 0
			}
			// Java: new RuleMatch(..., getCommaMessage(), messages.getString("double_commas_short"))
			rm := NewRuleMatch(r, sentence, fromPos, startPos+1, r.getCommaMessage())
			rm.ShortMessage = r.msg("double_commas_short", "")
			rm.SetSuggestedReplacement(commaChar)
			ruleMatches = append(ruleMatches, rm)
			commaCount = 0
		}
		if token != "." && token != commaChar {
			dotCount = 0
			commaCount = 0
		}
	}
	return ruleMatches
}
