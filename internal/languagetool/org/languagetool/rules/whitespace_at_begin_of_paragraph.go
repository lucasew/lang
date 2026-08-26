package rules

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// WhiteSpaceAtBeginOfParagraph ports org.languagetool.rules.WhiteSpaceAtBeginOfParagraph.
// Java: Categories.STYLE, ITSIssueType.Style; default ctor defaultActive=false → setDefaultOff();
// setOfficeDefaultOn() always.
type WhiteSpaceAtBeginOfParagraph struct {
	Messages        map[string]string
	Category        *Category
	IssueType       ITSIssueType
	DefaultOff      bool
	OfficeDefaultOn bool
}

// NewWhiteSpaceAtBeginOfParagraph ports WhiteSpaceAtBeginOfParagraph(ResourceBundle messages)
// → this(messages, false).
func NewWhiteSpaceAtBeginOfParagraph(messages map[string]string) *WhiteSpaceAtBeginOfParagraph {
	return NewWhiteSpaceAtBeginOfParagraphDefaultActive(messages, false)
}

// NewWhiteSpaceAtBeginOfParagraphDefaultActive ports
// WhiteSpaceAtBeginOfParagraph(ResourceBundle messages, boolean defaultActive).
func NewWhiteSpaceAtBeginOfParagraphDefaultActive(messages map[string]string, defaultActive bool) *WhiteSpaceAtBeginOfParagraph {
	r := &WhiteSpaceAtBeginOfParagraph{
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

func (r *WhiteSpaceAtBeginOfParagraph) GetID() string { return "WHITESPACE_PARAGRAPH_BEGIN" }

// GetDescription ports getDescription (whitespace_at_begin_parapgraph_desc — Java typo).
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

// IsOfficeDefaultOn ports Rule.isOfficeDefaultOn.
func (r *WhiteSpaceAtBeginOfParagraph) IsOfficeDefaultOn() bool {
	return r != nil && r.OfficeDefaultOn
}

// isWhitespaceDel ports WhiteSpaceAtBeginOfParagraph.isWhitespaceDel:
// whitespace that may be deleted; "\u200B" and linebreaks are excluded.
func (r *WhiteSpaceAtBeginOfParagraph) isWhitespaceDel(token *languagetool.AnalyzedTokenReadings) bool {
	// Java: token.isWhitespace() && !token.getToken().equals("\u200B") && !token.isLinebreak()
	return token.IsWhitespace() && token.GetToken() != "\u200B" && !token.IsLinebreak()
}

// Match ports WhiteSpaceAtBeginOfParagraph.match(AnalyzedSentence).
// Skips SENT_START (index 0); advances while leading deletable whitespace; if any was skipped
// and the next token is not a linebreak (and exists), match from tokens[1].startPos through
// tokens[i].endPos with suggestedReplacement = tokens[i].getToken().
func (r *WhiteSpaceAtBeginOfParagraph) Match(sentence *languagetool.AnalyzedSentence) []*RuleMatch {
	var ruleMatches []*RuleMatch
	tokens := sentence.GetTokens()
	// Java: for (i = 1; i < tokens.length && isWhitespaceDel(tokens[i]); i++);
	i := 1
	for i < len(tokens) && r.isWhitespaceDel(tokens[i]) {
		i++
	}
	if i > 1 && i < len(tokens) && !tokens[i].IsLinebreak() {
		msg := ""
		if r != nil && r.Messages != nil {
			msg = r.Messages["whitespace_at_begin_parapgraph_msg"]
		}
		if msg == "" {
			msg = "Don't start a paragraph with whitespace"
		}
		// Java: new RuleMatch(this, sentence, tokens[1].getStartPos(), tokens[i].getEndPos(), msg)
		rm := NewRuleMatch(r, sentence, tokens[1].GetStartPos(), tokens[i].GetEndPos(), msg)
		// Java: ruleMatch.setSuggestedReplacement(tokens[i].getToken())
		rm.SetSuggestedReplacement(tokens[i].GetToken())
		ruleMatches = append(ruleMatches, rm)
	}
	return ruleMatches
}
