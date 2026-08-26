package rules

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// WhiteSpaceBeforeParagraphEnd ports org.languagetool.rules.WhiteSpaceBeforeParagraphEnd.
// Java: Categories.STYLE, ITSIssueType.Style; default ctor defaultActive=false → setDefaultOff();
// setOfficeDefaultOn() always; minToCheckParagraph=0.
//
// SingleLineBreaksMarksPara ports Language.getSentenceTokenizer().singleLineBreaksMarksPara()
// as used by Tools.isParagraphEnd (Demo/SRX default false).
type WhiteSpaceBeforeParagraphEnd struct {
	Messages                  map[string]string
	SingleLineBreaksMarksPara bool
	Category                  *Category
	IssueType                 ITSIssueType
	DefaultOff                bool
	// OfficeDefaultOn ports Rule.setOfficeDefaultOn (Java always On for LO/OO).
	OfficeDefaultOn bool
}

// NewWhiteSpaceBeforeParagraphEnd ports WhiteSpaceBeforeParagraphEnd(ResourceBundle messages, Language lang)
// → this(messages, lang, false).
func NewWhiteSpaceBeforeParagraphEnd(messages map[string]string) *WhiteSpaceBeforeParagraphEnd {
	return NewWhiteSpaceBeforeParagraphEndDefaultActive(messages, false)
}

// NewWhiteSpaceBeforeParagraphEndDefaultActive ports
// WhiteSpaceBeforeParagraphEnd(ResourceBundle messages, Language lang, boolean defaultActive).
func NewWhiteSpaceBeforeParagraphEndDefaultActive(messages map[string]string, defaultActive bool) *WhiteSpaceBeforeParagraphEnd {
	r := &WhiteSpaceBeforeParagraphEnd{
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

func (r *WhiteSpaceBeforeParagraphEnd) GetID() string { return "WHITESPACE_PARAGRAPH" }

// GetDescription ports getDescription (whitespace_before_parapgraph_end_desc — Java typo).
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

// isParagraphEnd ports Tools.isParagraphEnd(sentences, n, lang) with
// lang.getSentenceTokenizer().singleLineBreaksMarksPara() = r.SingleLineBreaksMarksPara.
func (r *WhiteSpaceBeforeParagraphEnd) isParagraphEnd(sentences []*languagetool.AnalyzedSentence, nTest int) bool {
	return languagetool.IsParagraphEnd(sentences, nTest, r.SingleLineBreaksMarksPara)
}

// MatchList ports WhiteSpaceBeforeParagraphEnd.match(List<AnalyzedSentence>) bug-for-bug.
// For each paragraph-end sentence: skip trailing linebreaks (lb), then walk left through
// whitespace except U+200B (lw); if lw < lb, match from lw (or lw+1 when lw is whitespace)
// through tokens[lb].end with suggestedReplacement = tokens[lw].token when non-ws else "".
func (r *WhiteSpaceBeforeParagraphEnd) MatchList(sentences []*languagetool.AnalyzedSentence) []*RuleMatch {
	var ruleMatches []*RuleMatch
	pos := 0
	msg := ""
	if r != nil && r.Messages != nil {
		msg = r.Messages["whitespace_before_parapgraph_end_msg"]
	}
	if msg == "" {
		msg = "Don't end a paragraph with whitespace"
	}
	for n := 0; n < len(sentences); n++ {
		sentence := sentences[n]
		// Java: if (Tools.isParagraphEnd(sentences, n, lang))
		if r.isParagraphEnd(sentences, n) {
			tokens := sentence.GetTokens()
			// Java: for (lb = tokens.length - 1; lb > 0 && tokens[lb].isLinebreak(); lb--);
			lb := len(tokens) - 1
			for lb > 0 && tokens[lb].IsLinebreak() {
				lb--
			}
			// Java: for (lw = lb; lw > 0 && tokens[lw].isWhitespace() && !equals("\u200B"); lw--);
			lw := lb
			for lw > 0 && tokens[lw].IsWhitespace() && tokens[lw].GetToken() != "\u200B" {
				lw--
			}
			if lw < lb {
				// Java: fromPos = tokens[lw].isWhitespace() ? pos+tokens[lw+1].start : pos+tokens[lw].start
				var fromPos int
				if tokens[lw].IsWhitespace() {
					fromPos = pos + tokens[lw+1].GetStartPos()
				} else {
					fromPos = pos + tokens[lw].GetStartPos()
				}
				toPos := pos + tokens[lb].GetEndPos()
				rm := NewRuleMatch(r, sentence, fromPos, toPos, msg)
				// Java: if (lw > 0 && !tokens[lw].isWhitespace()) setSuggestedReplacement(tokens[lw].token) else ""
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
