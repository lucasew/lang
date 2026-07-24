package rules

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// LongParagraphRule ports org.languagetool.rules.LongParagraphRule.
// Java: STYLE, Style, setDefaultOff(), setOfficeDefaultOn(), Tag.picky; minToCheckParagraph=0.
type LongParagraphRule struct {
	Messages map[string]string
	MaxWords int
	// SingleLineBreaksMarksPara matches Demo/SRX default false → need \n\n
	SingleLineBreaksMarksPara bool
	Category                  *Category
	IssueType                 ITSIssueType
	DefaultOff                bool
	OfficeDefaultOn           bool
	// Tags ports Rule.tags (Java picky).
	Tags []Tag
}

func NewLongParagraphRule(messages map[string]string, maxWords int) *LongParagraphRule {
	return &LongParagraphRule{
		Messages:        messages,
		MaxWords:        maxWords,
		Category:        CatStyle.GetCategory(messages),
		IssueType:       ITSStyle,
		DefaultOff:      true,
		OfficeDefaultOn: true,
		Tags:            []Tag{TagPicky},
	}
}

func (r *LongParagraphRule) GetID() string { return "TOO_LONG_PARAGRAPH" }

// GetDescription ports MessageFormat(long_paragraph_rule_desc, maxWords).
func (r *LongParagraphRule) GetDescription() string {
	if r != nil && r.Messages != nil {
		if s := r.Messages["long_paragraph_rule_desc"]; s != "" {
			return messageFormat0(s, r.MaxWords)
		}
	}
	return fmt.Sprintf("Paragraph is too long (more than %d words)", r.MaxWords)
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

// IsOfficeDefaultOn ports Rule.isOfficeDefaultOn.
func (r *LongParagraphRule) IsOfficeDefaultOn() bool { return r != nil && r.OfficeDefaultOn }

// GetTags ports Rule.getTags (Java picky).
func (r *LongParagraphRule) GetTags() []Tag {
	if r == nil || len(r.Tags) == 0 {
		return nil
	}
	return append([]Tag(nil), r.Tags...)
}

// MinToCheckParagraph ports LongParagraphRule.minToCheckParagraph (Java returns 0).
func (r *LongParagraphRule) MinToCheckParagraph() int { return 0 }

func (r *LongParagraphRule) GetMessage() string {
	msg := ""
	if r.Messages != nil {
		msg = r.Messages["long_paragraph_rule_msg"]
	}
	if msg == "" {
		msg = "This paragraph is too long ({0} words)"
	}
	return messageFormat0(msg, r.MaxWords)
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
			// Java: !token.isWhitespace() && !token.isSentenceStart() && !token.isNonWord()
			if !token.IsWhitespace() && !token.IsSentenceStart() && !token.IsNonWord() {
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
