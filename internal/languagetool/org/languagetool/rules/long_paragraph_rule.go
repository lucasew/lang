package rules

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

var nonWordRE = regexp.MustCompile(`^[.?!…:;,~’'"„“”»«‚‘›‹()\[\]\-–—*×∗·+÷/=]$`)

// LongParagraphRule ports org.languagetool.rules.LongParagraphRule.
// Java: STYLE, Style, setDefaultOff(), Tag.picky.
type LongParagraphRule struct {
	Messages map[string]string
	MaxWords int
	// SingleLineBreaksMarksPara matches Demo/SRX default false → need \n\n
	SingleLineBreaksMarksPara bool
	Category                  *Category
	IssueType                 ITSIssueType
	DefaultOff                bool
	// Tags ports Rule.tags (Java picky).
	Tags []Tag
}

func NewLongParagraphRule(messages map[string]string, maxWords int) *LongParagraphRule {
	return &LongParagraphRule{
		Messages:   messages,
		MaxWords:   maxWords,
		Category:   CatStyle.GetCategory(messages),
		IssueType:  ITSStyle,
		DefaultOff: true,
		Tags:       []Tag{TagPicky},
	}
}

func (r *LongParagraphRule) GetID() string { return "TOO_LONG_PARAGRAPH" }

// GetDescription ports getDescription (long_paragraph_rule_desc).
func (r *LongParagraphRule) GetDescription() string {
	if r != nil && r.Messages != nil {
		if s := r.Messages["long_paragraph_rule_desc"]; s != "" {
			return s
		}
	}
	return "Paragraph is too long"
}

func (r *LongParagraphRule) GetCategory() *Category {
	if r == nil {
		return nil
	}
	return r.Category
}

func (r *LongParagraphRule) GetLocQualityIssueType() ITSIssueType {
	if r == nil || r.IssueType == "" {
		return ITSStyle
	}
	return r.IssueType
}

func (r *LongParagraphRule) IsDefaultOff() bool { return r != nil && r.DefaultOff }

func (r *LongParagraphRule) GetMessage() string {
	msg := r.Messages["long_paragraph_rule_msg"]
	if msg == "" {
		msg = "This paragraph is too long (%d words)"
	}
	if strings.Contains(msg, "%d") || strings.Contains(msg, "{0}") {
		return fmt.Sprintf(strings.ReplaceAll(msg, "{0}", "%d"), r.MaxWords)
	}
	return fmt.Sprintf(msg, r.MaxWords)
}

func (r *LongParagraphRule) isParagraphEnd(sentences []*languagetool.AnalyzedSentence, nTest int) bool {
	return languagetool.IsParagraphEnd(sentences, nTest, r.SingleLineBreaksMarksPara)
}

// MatchList ports match(List<AnalyzedSentence>).
func (r *LongParagraphRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*RuleMatch {
	var ruleMatches []*RuleMatch
	pos := 0
	startPos, endPos := 0, 0
	wordCount := 0
	paraHasLinebreaks := false
	for n := 0; n < len(sentences); n++ {
		sentence := sentences[n]
		paragraphEnd := r.isParagraphEnd(sentences, n)
		trimmed := regexp.MustCompile(`^\n+`).ReplaceAllString(sentence.GetText(), "")
		if !paragraphEnd && strings.Contains(trimmed, "\n") {
			paraHasLinebreaks = true
		}
		tokens := sentence.GetTokensWithoutWhitespace()
		for _, token := range tokens {
			if !token.IsWhitespace() && !token.IsSentenceStart() && !isNonWordToken(token) {
				wordCount++
				if wordCount == r.MaxWords {
					endPos = token.GetEndPos() + pos
				} else if wordCount == r.MaxWords-1 {
					startPos = token.GetStartPos() + pos
				}
			}
		}
		if paragraphEnd {
			if wordCount > r.MaxWords+5 && !paraHasLinebreaks {
				// last sentence of paragraph used for match attachment
				rm := NewRuleMatch(r, sentence, startPos, endPos, r.GetMessage())
				ruleMatches = append(ruleMatches, rm)
			}
			wordCount = 0
			paraHasLinebreaks = false
		}
		pos += sentence.GetCorrectedTextLength()
	}
	if wordCount > r.MaxWords {
		// no sentence ref in Java constructor without sentence - use last
		var last *languagetool.AnalyzedSentence
		if len(sentences) > 0 {
			last = sentences[len(sentences)-1]
		}
		rm := NewRuleMatch(r, last, startPos, endPos, r.GetMessage())
		ruleMatches = append(ruleMatches, rm)
	}
	return ruleMatches
}

func isNonWordToken(token *languagetool.AnalyzedTokenReadings) bool {
	return nonWordRE.MatchString(token.GetToken())
}
