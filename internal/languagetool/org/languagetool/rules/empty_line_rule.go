package rules

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// EmptyLineRule ports org.languagetool.rules.EmptyLineRule.
// Java: STYLE, Style; default ctor setDefaultOff (defaultActive=false); setOfficeDefaultOn(); minToCheckParagraph=1.
//
// SingleLineBreaksMarksPara ports Language.getSentenceTokenizer().singleLineBreaksMarksPara()
// as used by Tools.isParagraphEnd and isSecondParagraphEndMark (Demo/SRX default false).
type EmptyLineRule struct {
	Messages                  map[string]string
	SingleLineBreaksMarksPara bool
	Category                  *Category
	IssueType                 ITSIssueType
	DefaultOff                bool
	// OfficeDefaultOn ports Rule.setOfficeDefaultOn (Java always On for LO/OO).
	OfficeDefaultOn bool
}

// NewEmptyLineRule ports EmptyLineRule(ResourceBundle messages, Language lang)
// → defaultActive false → setDefaultOff(); setOfficeDefaultOn().
func NewEmptyLineRule(messages map[string]string) *EmptyLineRule {
	return NewEmptyLineRuleDefaultActive(messages, false)
}

// NewEmptyLineRuleDefaultActive ports
// EmptyLineRule(ResourceBundle messages, Language lang, boolean defaultActive).
func NewEmptyLineRuleDefaultActive(messages map[string]string, defaultActive bool) *EmptyLineRule {
	r := &EmptyLineRule{
		Messages:        messages,
		Category:        CatStyle.GetCategory(messages),
		IssueType:       ITSStyle,
		OfficeDefaultOn: true, // Java setOfficeDefaultOn() always
	}
	if !defaultActive {
		r.DefaultOff = true // Java setDefaultOff()
	}
	return r
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

// isParagraphEnd ports Tools.isParagraphEnd(sentences, n, lang) with
// lang.getSentenceTokenizer().singleLineBreaksMarksPara() = r.SingleLineBreaksMarksPara.
func (r *EmptyLineRule) isParagraphEnd(sentences []*languagetool.AnalyzedSentence, nTest int) bool {
	return languagetool.IsParagraphEnd(sentences, nTest, r.SingleLineBreaksMarksPara)
}

// isSecondParagraphEndMark ports EmptyLineRule.isSecondParagraphEndMark.
func (r *EmptyLineRule) isSecondParagraphEndMark(sentence string) bool {
	if r.SingleLineBreaksMarksPara {
		// Java: endsWith("\n\n") || endsWith("\n\r\n\r")
		return strings.HasSuffix(sentence, "\n\n") || strings.HasSuffix(sentence, "\n\r\n\r")
	}
	// Java: endsWith("\n\n\n\n") || endsWith("\n\r\n\r\n\r\n\r") || endsWith("\r\n\r\n\r\n\r\n")
	return strings.HasSuffix(sentence, "\n\n\n\n") ||
		strings.HasSuffix(sentence, "\n\r\n\r\n\r\n\r") ||
		strings.HasSuffix(sentence, "\r\n\r\n\r\n\r\n")
}

// MatchList ports EmptyLineRule.match(List<AnalyzedSentence>) bug-for-bug.
// Loop n in 0..size-2; on paragraph end + second end mark + tokensWithoutWhitespace length > 1,
// match last non-ws token span with pos accumulation via getCorrectedTextLength.
func (r *EmptyLineRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*RuleMatch {
	var ruleMatches []*RuleMatch
	pos := 0
	msg := ""
	if r != nil && r.Messages != nil {
		msg = r.Messages["empty_line_rule_msg"]
	}
	if msg == "" {
		msg = "Empty line"
	}
	for n := 0; n < len(sentences)-1; n++ {
		sentence := sentences[n]
		// Java: if (Tools.isParagraphEnd(...)) { if (isSecondParagraphEndMark(...)) { ... } }
		if r.isParagraphEnd(sentences, n) && r.isSecondParagraphEndMark(sentence.GetText()) {
			tokens := sentence.GetTokensWithoutWhitespace()
			if len(tokens) > 1 {
				last := tokens[len(tokens)-1]
				fromPos := pos + last.GetStartPos()
				toPos := pos + last.GetEndPos()
				ruleMatches = append(ruleMatches, NewRuleMatch(r, sentence, fromPos, toPos, msg))
			}
		}
		pos += sentence.GetCorrectedTextLength()
	}
	return ruleMatches
}
