package rules

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// WhiteSpaceBeforeParagraphEnd ports org.languagetool.rules.WhiteSpaceBeforeParagraphEnd.
// Java: STYLE, Style; default ctor setDefaultOff; setOfficeDefaultOn; minToCheckParagraph=0.
type WhiteSpaceBeforeParagraphEnd struct {
	Messages                  map[string]string
	SingleLineBreaksMarksPara bool
	Category                  *Category
	IssueType                 ITSIssueType
	DefaultOff                bool
	OfficeDefaultOn           bool
}

func NewWhiteSpaceBeforeParagraphEnd(messages map[string]string) *WhiteSpaceBeforeParagraphEnd {
	// Java (messages, lang) → defaultActive false → setDefaultOff(); setOfficeDefaultOn().
	return &WhiteSpaceBeforeParagraphEnd{
		Messages:        messages,
		Category:        CatStyle.GetCategory(messages),
		IssueType:       ITSStyle,
		DefaultOff:      true,
		OfficeDefaultOn: true,
	}
}

func (r *WhiteSpaceBeforeParagraphEnd) GetID() string { return "WHITESPACE_PARAGRAPH" }

// GetDescription ports getDescription (whitespace_before_parapgraph_end_desc).
func (r *WhiteSpaceBeforeParagraphEnd) GetDescription() string {
	if r != nil && r.Messages != nil {
		if s := r.Messages["whitespace_before_parapgraph_end_desc"]; s != "" {
			return s
		}
	}
	return "Whitespace before paragraph end"
}

func (r *WhiteSpaceBeforeParagraphEnd) GetCategory() *Category {
	if r == nil {
		return nil
	}
	return r.Category
}

func (r *WhiteSpaceBeforeParagraphEnd) GetLocQualityIssueType() ITSIssueType {
	if r == nil || r.IssueType == "" {
		return ITSStyle
	}
	return r.IssueType
}

func (r *WhiteSpaceBeforeParagraphEnd) IsDefaultOff() bool { return r != nil && r.DefaultOff }

// IsOfficeDefaultOn ports Rule.isOfficeDefaultOn.
func (r *WhiteSpaceBeforeParagraphEnd) IsOfficeDefaultOn() bool {
	return r != nil && r.OfficeDefaultOn
}

// MinToCheckParagraph ports WhiteSpaceBeforeParagraphEnd.minToCheckParagraph (Java returns 0).
func (r *WhiteSpaceBeforeParagraphEnd) MinToCheckParagraph() int { return 0 }

func (r *WhiteSpaceBeforeParagraphEnd) isParagraphEnd(sentences []*languagetool.AnalyzedSentence, nTest int) bool {
	return languagetool.IsParagraphEnd(sentences, nTest, r.SingleLineBreaksMarksPara)
}

// MatchList ports match(List<AnalyzedSentence>).
func (r *WhiteSpaceBeforeParagraphEnd) MatchList(sentences []*languagetool.AnalyzedSentence) []*RuleMatch {
	var ruleMatches []*RuleMatch
	pos := 0
	msg := r.Messages["whitespace_before_parapgraph_end_msg"]
	if msg == "" {
		msg = "Don't end a paragraph with whitespace"
	}
	for n := 0; n < len(sentences); n++ {
		sentence := sentences[n]
		if r.isParagraphEnd(sentences, n) {
			tokens := sentence.GetTokens()
			lb := len(tokens) - 1
			for lb > 0 && tokens[lb].IsLinebreak() {
				lb--
			}
			lw := lb
			for lw > 0 && tokens[lw].IsWhitespace() && tokens[lw].GetToken() != "\u200B" {
				lw--
			}
			if lw < lb {
				fromPos := pos + tokens[lw].GetStartPos()
				if tokens[lw].IsWhitespace() && lw+1 < len(tokens) {
					fromPos = pos + tokens[lw+1].GetStartPos()
				}
				toPos := pos + tokens[lb].GetEndPos()
				rm := NewRuleMatch(r, sentence, fromPos, toPos, msg)
				if lw > 0 && !tokens[lw].IsWhitespace() {
					rm.SetSuggestedReplacement(tokens[lw].GetToken())
				} else {
					rm.SetSuggestedReplacement("")
				}
				ruleMatches = append(ruleMatches, rm)
			}
		}
		pos += sentence.GetCorrectedTextLength()
	}
	return ruleMatches
}
