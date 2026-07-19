package rules

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// WhiteSpaceAtBeginOfParagraph ports org.languagetool.rules.WhiteSpaceAtBeginOfParagraph.
// Java: STYLE, Style; default ctor setDefaultOff.
type WhiteSpaceAtBeginOfParagraph struct {
	Messages   map[string]string
	Category   *Category
	IssueType  ITSIssueType
	DefaultOff bool
}

func NewWhiteSpaceAtBeginOfParagraph(messages map[string]string) *WhiteSpaceAtBeginOfParagraph {
	return &WhiteSpaceAtBeginOfParagraph{
		Messages:   messages,
		Category:   CatStyle.GetCategory(messages),
		IssueType:  ITSStyle,
		DefaultOff: true,
	}
}

func (r *WhiteSpaceAtBeginOfParagraph) GetID() string { return "WHITESPACE_PARAGRAPH_BEGIN" }

// GetDescription ports getDescription (whitespace_at_begin_parapgraph_desc).
func (r *WhiteSpaceAtBeginOfParagraph) GetDescription() string {
	if r != nil && r.Messages != nil {
		if s := r.Messages["whitespace_at_begin_parapgraph_desc"]; s != "" {
			return s
		}
	}
	return "Whitespace at begin of paragraph"
}

func (r *WhiteSpaceAtBeginOfParagraph) GetCategory() *Category {
	if r == nil {
		return nil
	}
	return r.Category
}

func (r *WhiteSpaceAtBeginOfParagraph) GetLocQualityIssueType() ITSIssueType {
	if r == nil || r.IssueType == "" {
		return ITSStyle
	}
	return r.IssueType
}

func (r *WhiteSpaceAtBeginOfParagraph) IsDefaultOff() bool { return r != nil && r.DefaultOff }

func isWhitespaceDel(token *languagetool.AnalyzedTokenReadings) bool {
	return token.IsWhitespace() && token.GetToken() != "\u200B" && !token.IsLinebreak()
}

func (r *WhiteSpaceAtBeginOfParagraph) Match(sentence *languagetool.AnalyzedSentence) []*RuleMatch {
	var ruleMatches []*RuleMatch
	tokens := sentence.GetTokens()
	i := 1
	for i < len(tokens) && isWhitespaceDel(tokens[i]) {
		i++
	}
	if i > 1 && i < len(tokens) && !tokens[i].IsLinebreak() {
		msg := r.Messages["whitespace_at_begin_parapgraph_msg"]
		if msg == "" {
			msg = "Don't start a paragraph with whitespace"
		}
		rm := NewRuleMatch(r, sentence, tokens[1].GetStartPos(), tokens[i].GetEndPos(), msg)
		rm.SetSuggestedReplacement(tokens[i].GetToken())
		ruleMatches = append(ruleMatches, rm)
	}
	return ruleMatches
}
