package rules

import (
	"regexp"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// PunctuationMarkAtParagraphEnd2 ports org.languagetool.rules.PunctuationMarkAtParagraphEnd2.
// Java: PUNCTUATION, Grammar; setDefaultOff(); minToCheckParagraph=0.
type PunctuationMarkAtParagraphEnd2 struct {
	Messages                  map[string]string
	SingleLineBreaksMarksPara bool
	Category                  *Category
	IssueType                 ITSIssueType
	DefaultOff                bool
}

// more than this many word tokens needed for a "real" paragraph
const paraEnd2TokenThreshold = 10

var paraEnd2FinalPunct = regexp.MustCompile(`^[:.?!…]$`)

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

// MinToCheckParagraph ports minToCheckParagraph (Java returns 0).
func (r *PunctuationMarkAtParagraphEnd2) MinToCheckParagraph() int { return 0 }

func (r *PunctuationMarkAtParagraphEnd2) isParagraphEnd(sentences []*languagetool.AnalyzedSentence, nTest int) bool {
	return languagetool.IsParagraphEnd(sentences, nTest, r.SingleLineBreaksMarksPara)
}

func getLastNonSpaceToken(tokens []*languagetool.AnalyzedTokenReadings) *languagetool.AnalyzedTokenReadings {
	for i := len(tokens) - 1; i >= 0; i-- {
		if !tokens[i].IsWhitespace() {
			return tokens[i]
		}
	}
	return nil
}

// MatchList ports match(List<AnalyzedSentence>).
func (r *PunctuationMarkAtParagraphEnd2) MatchList(sentences []*languagetool.AnalyzedSentence) []*RuleMatch {
	var ruleMatches []*RuleMatch
	pos := 0
	tokenCount := 0
	for sentPos, sentence := range sentences {
		tokens := sentence.GetTokens()
		for _, token := range tokens {
			if !token.IsNonWord() && !token.IsWhitespace() {
				tokenCount++
			}
		}
		lastNonSpace := getLastNonSpaceToken(tokens)
		isParaEnd := r.isParagraphEnd(sentences, sentPos)
		if isParaEnd && tokenCount > paraEnd2TokenThreshold &&
			lastNonSpace != nil &&
			!paraEnd2FinalPunct.MatchString(lastNonSpace.GetToken()) &&
			!lastNonSpace.IsNonWord() {
			msg := "Add a punctuation mark at paragraph end"
			if r.Messages != nil {
				if m := r.Messages["punctuation_mark_paragraph_end_msg"]; m != "" {
					msg = m
				}
			}
			from := pos + lastNonSpace.GetStartPos()
			to := pos + lastNonSpace.GetEndPos()
			rm := NewRuleMatch(r, sentence, from, to, msg)
			rm.SetSuggestedReplacement(lastNonSpace.GetToken() + ".")
			ruleMatches = append(ruleMatches, rm)
		}
		if isParaEnd {
			tokenCount = 0
		}
		pos += sentence.GetCorrectedTextLength()
	}
	return ruleMatches
}
