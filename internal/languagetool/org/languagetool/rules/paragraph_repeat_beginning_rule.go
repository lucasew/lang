package rules

import (
	"regexp"
	"unicode"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// ParagraphRepeatBeginningRule ports org.languagetool.rules.ParagraphRepeatBeginningRule.
//
// Java ctor: super(messages); Categories.STYLE; this.lang = lang; ITSIssueType.Style; setDefaultOff().
// Language is used only for Tools.isParagraphEnd → singleLineBreaksMarksPara (field below).
//
// TextLevelRule; STYLE; Style; setDefaultOff(); minToCheckParagraph=1.
// id: PARAGRAPH_REPEAT_BEGINNING_RULE
// messages: repetition_paragraph_beginning_desc / repetition_paragraph_beginning_last_msg
type ParagraphRepeatBeginningRule struct {
	Messages map[string]string
	// Category ports Rule.category (Java STYLE).
	Category *Category
	// IssueType ports getLocQualityIssueType (Java Style).
	IssueType ITSIssueType
	// DefaultOff ports setDefaultOff (Java true).
	DefaultOff bool
	// IsArticle optional override of isArticle(AnalyzedTokenReadings); when nil,
	// Java default: hasPosTagStartingWith("DT"). German subclass sets ART.
	IsArticle func(token *languagetool.AnalyzedTokenReadings) bool
	// SingleLineBreaksMarksPara ports Language.getSentenceTokenizer().singleLineBreaksMarksPara()
	// as used by Tools.isParagraphEnd (Demo/SRX default false → need \n\n or next starts with \n).
	SingleLineBreaksMarksPara bool
	// RuleID optional override of getId (German subclass returns GERMAN_...).
	RuleID string
}

// QUOTES_REGEX ports ParagraphRepeatBeginningRule.QUOTES_REGEX.
// Java Pattern.matcher(token).matches() requires the whole token to be one quote/bracket char.
// Anchored so Go MatchString matches Java matches(), not find().
var paraQuotesRE = regexp.MustCompile(`\A[’'"„“”»«‚‘›‹()\[\]]\z`)

// NewParagraphRepeatBeginningRule ports
// ParagraphRepeatBeginningRule(ResourceBundle messages, Language lang).
// Caller sets SingleLineBreaksMarksPara from the language sentence tokenizer when needed.
func NewParagraphRepeatBeginningRule(messages map[string]string) *ParagraphRepeatBeginningRule {
	return &ParagraphRepeatBeginningRule{
		Messages:   messages,
		Category:   CatStyle.GetCategory(messages),
		IssueType:  ITSStyle,
		DefaultOff: true,
	}
}

func (r *ParagraphRepeatBeginningRule) GetID() string {
	if r.RuleID != "" {
		return r.RuleID
	}
	return "PARAGRAPH_REPEAT_BEGINNING_RULE"
}

// GetDescription ports getDescription (repetition_paragraph_beginning_desc).
func (r *ParagraphRepeatBeginningRule) GetDescription() string {
	if r != nil && r.Messages != nil {
		if s := r.Messages["repetition_paragraph_beginning_desc"]; s != "" {
			return s
		}
	}
	return "Same beginning of paragraph"
}

// GetCategory ports Rule.getCategory.
func (r *ParagraphRepeatBeginningRule) GetCategory() *Category {
	if r == nil {
		return nil
	}
	return r.Category
}

// GetLocQualityIssueType ports Rule.getLocQualityIssueType.
func (r *ParagraphRepeatBeginningRule) GetLocQualityIssueType() ITSIssueType {
	if r == nil || r.IssueType == "" {
		return ITSStyle
	}
	return r.IssueType
}

// IsDefaultOff ports Rule.isDefaultOff.
func (r *ParagraphRepeatBeginningRule) IsDefaultOff() bool { return r != nil && r.DefaultOff }

// MinToCheckParagraph ports minToCheckParagraph (Java returns 1).
func (r *ParagraphRepeatBeginningRule) MinToCheckParagraph() int { return 1 }

// isParagraphEnd ports Tools.isParagraphEnd(sentences, n, lang) via SingleLineBreaksMarksPara.
func (r *ParagraphRepeatBeginningRule) isParagraphEnd(sentences []*languagetool.AnalyzedSentence, nTest int) bool {
	return languagetool.IsParagraphEnd(sentences, nTest, r.SingleLineBreaksMarksPara)
}

// isArticle ports ParagraphRepeatBeginningRule.isArticle (overridable via IsArticle field).
// Java default: token.hasPosTagStartingWith("DT").
func (r *ParagraphRepeatBeginningRule) isArticle(token *languagetool.AnalyzedTokenReadings) bool {
	if r.IsArticle != nil {
		return r.IsArticle(token)
	}
	return token != nil && token.HasPosTagStartingWith("DT")
}

// numCharEqualBeginning ports private numCharEqualBeginning.
// Returns lastTokens[nToken].getEndPos() when paragraph beginnings are equal, else 0.
func (r *ParagraphRepeatBeginningRule) numCharEqualBeginning(lastTokens, nextTokens []*languagetool.AnalyzedTokenReadings) int {
	if len(lastTokens) < 2 || len(nextTokens) < 2 ||
		lastTokens[1].IsWhitespace() || nextTokens[1].IsWhitespace() {
		return 0
	}
	nToken := 1
	lastToken := lastTokens[nToken].GetToken()
	nextToken := nextTokens[nToken].GetToken()
	// Java: QUOTES_REGEX.matcher(lastToken).matches() && lastToken.equals(nextToken)
	if paraQuotesRE.MatchString(lastToken) && lastToken == nextToken {
		if len(lastTokens) <= nToken+1 || len(nextTokens) <= nToken+1 {
			return 0
		}
		nToken++
		lastToken = lastTokens[nToken].GetToken()
		nextToken = nextTokens[nToken].GetToken()
	}
	// Java: Character.isLetter(lastToken.charAt(0)) — first UTF-16 code unit
	if lastToken == "" || !javaIsLetterCharAt0(lastToken) {
		return 0
	}
	if len(lastTokens) > nToken+1 && r.isArticle(lastTokens[nToken]) && lastToken == nextToken {
		if len(nextTokens) <= nToken+1 {
			return 0
		}
		nToken++
		lastToken = lastTokens[nToken].GetToken()
		nextToken = nextTokens[nToken].GetToken()
	}
	if lastToken == "" || !javaIsLetterCharAt0(lastToken) {
		return 0
	}
	if lastToken == nextToken {
		return lastTokens[nToken].GetEndPos()
	}
	return 0
}

// MatchList ports match(List<AnalyzedSentence>).
//
// Java position accumulation uses sentences.get(n).getText().length() (UTF-16), NOT
// getCorrectedTextLength — preserve that quirk bug-for-bug.
func (r *ParagraphRepeatBeginningRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*RuleMatch {
	var ruleMatches []*RuleMatch
	if len(sentences) < 1 {
		return ruleMatches
	}
	nextPos := 0
	lastPos := 0
	lastSentence := sentences[0]
	lastTokens := lastSentence.GetTokensWithoutWhitespace()
	msg := ""
	if r.Messages != nil {
		msg = r.Messages["repetition_paragraph_beginning_last_msg"]
	}
	if msg == "" {
		msg = "Same beginning as last paragraph"
	}
	for n := 0; n < len(sentences)-1; n++ {
		// Java: nextPos += sentences.get(n).getText().length() (String.length = UTF-16 units)
		nextPos += utf16Len(sentences[n].GetText())
		if !r.isParagraphEnd(sentences, n) {
			continue
		}
		nextSentence := sentences[n+1]
		nextTokens := nextSentence.GetTokensWithoutWhitespace()
		endPos := r.numCharEqualBeginning(lastTokens, nextTokens)
		if endPos > 0 {
			startPos := lastPos + lastTokens[1].GetStartPos()
			if startPos < lastPos+endPos {
				rm := NewRuleMatch(r, lastSentence, startPos, lastPos+endPos, msg)
				ruleMatches = append(ruleMatches, rm)

				// Java reuses endPos from lastTokens on the next-sentence match (quirk).
				startPos2 := nextPos + nextTokens[1].GetStartPos()
				rm2 := NewRuleMatch(r, nextSentence, startPos2, nextPos+endPos, msg)
				ruleMatches = append(ruleMatches, rm2)
			}
		}
		// Java only advances last* inside the paragraph-end branch.
		lastSentence = nextSentence
		lastTokens = nextTokens
		lastPos = nextPos
	}
	return ruleMatches
}

// javaIsLetterCharAt0 ports Character.isLetter(str.charAt(0)).
func javaIsLetterCharAt0(s string) bool {
	u := utf16.Encode([]rune(s))
	if len(u) == 0 {
		return false
	}
	return unicode.IsLetter(rune(u[0]))
}
