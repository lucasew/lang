package rules

import (
	"regexp"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// PunctuationMarkAtParagraphEnd2 ports org.languagetool.rules.PunctuationMarkAtParagraphEnd2.
// Simplified implementation of PunctuationMarkAtParagraphEnd.
// Java: Categories.PUNCTUATION, ITSIssueType.Grammar; always setDefaultOff();
// minToCheckParagraph=0. Language is represented by SingleLineBreaksMarksPara for
// Tools.isParagraphEnd (Demo/SRX SimpleSentenceTokenizer default false → need \n\n
// or next sentence starting with newline; last sentence is always a para end).
type PunctuationMarkAtParagraphEnd2 struct {
	Messages map[string]string
	// SingleLineBreaksMarksPara ports Language.getSentenceTokenizer().singleLineBreaksMarksPara()
	// as used by Tools.isParagraphEnd.
	SingleLineBreaksMarksPara bool
	Category                  *Category
	IssueType                 ITSIssueType
	DefaultOff                bool
}

// more than this many word tokens needed for a "real" paragraph that requires a period (etc) at the end:
// Java TOKEN_THRESHOLD = 10
const paraEnd2TokenThreshold = 10

// Java: lastNonSpaceToken.getToken().matches("[:.?!…]") — String.matches is full-string.
var paraEnd2FinalPunct = regexp.MustCompile(`^[:.?!…]$`)

// NewPunctuationMarkAtParagraphEnd2 ports
// PunctuationMarkAtParagraphEnd2(ResourceBundle messages, Language lang).
// Language for Tools.isParagraphEnd is carried as SingleLineBreaksMarksPara (zero value
// matches Demo: false). Always setDefaultOff().
func NewPunctuationMarkAtParagraphEnd2(messages map[string]string) *PunctuationMarkAtParagraphEnd2 {
	return &PunctuationMarkAtParagraphEnd2{
		Messages:   messages,
		Category:   CatPunctuation.GetCategory(messages),
		IssueType:  ITSGrammar,
		DefaultOff: true,
	}
}

func (r *PunctuationMarkAtParagraphEnd2) GetID() string { return "PUNCTUATION_PARAGRAPH_END2" }

// GetDescription ports getDescription (punctuation_mark_paragraph_end_desc).
func (r *PunctuationMarkAtParagraphEnd2) GetDescription() string {
	if r != nil && r.Messages != nil {
		if s := r.Messages["punctuation_mark_paragraph_end_desc"]; s != "" {
			return s
		}
	}
	return "No punctuation mark at the end of paragraph"
}

func (r *PunctuationMarkAtParagraphEnd2) GetCategory() *Category {
	if r == nil {
		return nil
	}
	return r.Category
}

func (r *PunctuationMarkAtParagraphEnd2) GetLocQualityIssueType() ITSIssueType {
	if r == nil || r.IssueType == "" {
		return ITSGrammar
	}
	return r.IssueType
}

func (r *PunctuationMarkAtParagraphEnd2) IsDefaultOff() bool { return r != nil && r.DefaultOff }

// SetDefaultOn ports Rule.setDefaultOn (Java test enables the default-off rule).
func (r *PunctuationMarkAtParagraphEnd2) SetDefaultOn() {
	if r != nil {
		r.DefaultOff = false
	}
}

// SetDefaultOff ports Rule.setDefaultOff.
func (r *PunctuationMarkAtParagraphEnd2) SetDefaultOff() {
	if r != nil {
		r.DefaultOff = true
	}
}

// MinToCheckParagraph ports minToCheckParagraph (Java returns 0).
func (r *PunctuationMarkAtParagraphEnd2) MinToCheckParagraph() int { return 0 }

// EstimateContextForSureMatch ports TextLevelRule (Java always -1).
func (r *PunctuationMarkAtParagraphEnd2) EstimateContextForSureMatch() int { return -1 }

func (r *PunctuationMarkAtParagraphEnd2) isParagraphEnd(sentences []*languagetool.AnalyzedSentence, nTest int) bool {
	// Java: Tools.isParagraphEnd(sentences, sentPos, lang)
	return languagetool.IsParagraphEnd(sentences, nTest, r.SingleLineBreaksMarksPara)
}

// getLastNonSpaceToken ports private getLastNonSpaceToken(AnalyzedTokenReadings[]).
func getLastNonSpaceToken(tokens []*languagetool.AnalyzedTokenReadings) *languagetool.AnalyzedTokenReadings {
	for i := len(tokens) - 1; i >= 0; i-- {
		if !tokens[i].IsWhitespace() {
			return tokens[i]
		}
	}
	return nil
}

// MatchList ports match(List<AnalyzedSentence>) bug-for-bug with Java control flow:
// accumulate non-nonWord non-whitespace tokens until paragraph end; if count >
// TOKEN_THRESHOLD and last non-space token is not [:.?!…] and not isNonWord, match.
func (r *PunctuationMarkAtParagraphEnd2) MatchList(sentences []*languagetool.AnalyzedSentence) []*RuleMatch {
	var ruleMatches []*RuleMatch
	pos := 0
	// sentPos is the range index (Java for-each with sentPos++)
	tokenCount := 0
	for sentPos, sentence := range sentences {
		tokens := sentence.GetTokens()
		for _, token := range tokens {
			if !token.IsNonWord() && !token.IsWhitespace() {
				tokenCount++
			}
		}
		lastNonSpaceToken := getLastNonSpaceToken(tokens)
		isParaEnd := r.isParagraphEnd(sentences, sentPos)
		if isParaEnd && tokenCount > paraEnd2TokenThreshold &&
			lastNonSpaceToken != nil &&
			!paraEnd2FinalPunct.MatchString(lastNonSpaceToken.GetToken()) &&
			!lastNonSpaceToken.IsNonWord() {
			msg := "Please add a punctuation mark at the end of paragraph."
			if r.Messages != nil {
				if m := r.Messages["punctuation_mark_paragraph_end_msg"]; m != "" {
					msg = m
				}
			}
			from := pos + lastNonSpaceToken.GetStartPos()
			to := pos + lastNonSpaceToken.GetEndPos()
			rm := NewRuleMatch(r, sentence, from, to, msg)
			rm.SetSuggestedReplacement(lastNonSpaceToken.GetToken() + ".")
			ruleMatches = append(ruleMatches, rm)
		}
		if isParaEnd {
			tokenCount = 0
		}
		// Java: sentPos++ then pos += sentence.getCorrectedTextLength()
		pos += sentence.GetCorrectedTextLength()
	}
	return ruleMatches
}
