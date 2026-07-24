package rules

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// EmptyLineRule ports org.languagetool.rules.EmptyLineRule.
// Java: STYLE, Style; default ctor setDefaultOff (defaultActive=false); setOfficeDefaultOn(); minToCheckParagraph=1.
type EmptyLineRule struct {
	Messages                  map[string]string
	SingleLineBreaksMarksPara bool
	Category                  *Category
	IssueType                 ITSIssueType
	DefaultOff                bool
	// OfficeDefaultOn ports Rule.setOfficeDefaultOn (Java always On for LO/OO).
	OfficeDefaultOn bool
}

func NewEmptyLineRule(messages map[string]string) *EmptyLineRule {
	// Java EmptyLineRule(messages, lang) → defaultActive false → setDefaultOff(); setOfficeDefaultOn().
	return &EmptyLineRule{
		Messages:        messages,
		Category:        CatStyle.GetCategory(messages),
		IssueType:       ITSStyle,
		DefaultOff:      true,
		OfficeDefaultOn: true,
	}
}

func (r *EmptyLineRule) GetID() string { return "EMPTY_LINE" }

// GetDescription ports EmptyLineRule.getDescription (empty_line_rule_desc).
func (r *EmptyLineRule) GetDescription() string {
	if r != nil && r.Messages != nil {
		if s := r.Messages["empty_line_rule_desc"]; s != "" {
			return s
		}
	}
	return "Empty line"
}

func (r *EmptyLineRule) GetCategory() *Category {
	if r == nil {
		return nil
	}
	return r.Category
}

func (r *EmptyLineRule) GetLocQualityIssueType() ITSIssueType {
	if r == nil || r.IssueType == "" {
		return ITSStyle
	}
	return r.IssueType
}

func (r *EmptyLineRule) IsDefaultOff() bool { return r != nil && r.DefaultOff }

// IsOfficeDefaultOn ports Rule.isOfficeDefaultOn.
func (r *EmptyLineRule) IsOfficeDefaultOn() bool { return r != nil && r.OfficeDefaultOn }

// MinToCheckParagraph ports EmptyLineRule.minToCheckParagraph (Java returns 1).
func (r *EmptyLineRule) MinToCheckParagraph() int { return 1 }

func (r *EmptyLineRule) isParagraphEnd(sentences []*languagetool.AnalyzedSentence, nTest int) bool {
	return languagetool.IsParagraphEnd(sentences, nTest, r.SingleLineBreaksMarksPara)
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
