package rules

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// PunctuationMarkAtParagraphEnd ports org.languagetool.rules.PunctuationMarkAtParagraphEnd.
// Java: PUNCTUATION, Grammar; default ctor defaultActive=true (not default-off); Tag.picky.
type PunctuationMarkAtParagraphEnd struct {
	Messages map[string]string
	// SingleLineBreaksMarksPara matches Demo/SRX default false → need \n\n
	SingleLineBreaksMarksPara bool
	Category                  *Category
	IssueType                 ITSIssueType
	DefaultOff                bool
	// Tags ports Rule.tags (Java picky).
	Tags []Tag
}

var (
	paraEndPunctMarks = []string{".", "!", "?", ":", ",", ";"}
	paraEndQuoteMarks = []string{"„", "»", "«", "\"", "”", "″", "’", "‚", "‘", "›", "‹", "′", "'"}
	paraEndNumericRE  = regexp.MustCompile(`^[0-9.]+$`)
)

const maxURLLength = 30

func NewPunctuationMarkAtParagraphEnd(messages map[string]string) *PunctuationMarkAtParagraphEnd {
	// Java (messages, lang) → this(messages, lang, true) → defaultActive true → NOT setDefaultOff.
	return &PunctuationMarkAtParagraphEnd{
		Messages:   messages,
		Category:   CatPunctuation.GetCategory(messages),
		IssueType:  ITSGrammar,
		DefaultOff: false,
		Tags:       []Tag{TagPicky},
	}
}

func (r *PunctuationMarkAtParagraphEnd) GetID() string { return "PUNCTUATION_PARAGRAPH_END" }

// GetDescription ports getDescription (punctuation_mark_paragraph_end_desc).
func (r *PunctuationMarkAtParagraphEnd) GetDescription() string {
	if r != nil && r.Messages != nil {
		if s := r.Messages["punctuation_mark_paragraph_end_desc"]; s != "" {
			return s
		}
	}
	return "No punctuation mark at the end of paragraph"
}

func (r *PunctuationMarkAtParagraphEnd) GetCategory() *Category {
	if r == nil {
		return nil
	}
	return r.Category
}

func (r *PunctuationMarkAtParagraphEnd) GetLocQualityIssueType() ITSIssueType {
	if r == nil || r.IssueType == "" {
		return ITSGrammar
	}
	return r.IssueType
}

func (r *PunctuationMarkAtParagraphEnd) IsDefaultOff() bool { return r != nil && r.DefaultOff }

// GetTags ports Rule.getTags (Java Tag.picky).
func (r *PunctuationMarkAtParagraphEnd) GetTags() []Tag {
	if r == nil || len(r.Tags) == 0 {
		return nil
	}
	return append([]Tag(nil), r.Tags...)
}

// MinToCheckParagraph ports minToCheckParagraph (Java returns 0).
func (r *PunctuationMarkAtParagraphEnd) MinToCheckParagraph() int { return 0 }

func (r *PunctuationMarkAtParagraphEnd) isParagraphEnd(sentences []*languagetool.AnalyzedSentence, nTest int) bool {
	return languagetool.IsParagraphEnd(sentences, nTest, r.SingleLineBreaksMarksPara)
}

func stringEqualsAny(token string, any []string) bool {
	for _, s := range any {
		if token == s {
			return true
		}
	}
	return false
}

func isParaQuotationMark(tk *languagetool.AnalyzedTokenReadings) bool {
	return stringEqualsAny(tk.GetToken(), paraEndQuoteMarks)
}

func isParaPunctuationMark(tk *languagetool.AnalyzedTokenReadings) bool {
	return stringEqualsAny(tk.GetToken(), paraEndPunctMarks)
}

func isParaWord(tk *languagetool.AnalyzedTokenReadings) bool {
	tok := tk.GetToken()
	if tok == "" {
		return false
	}
	for _, r := range tok {
		return unicode.IsLetter(r)
	}
	return false
}

func isParaNumeric(s string) bool {
	return paraEndNumericRE.MatchString(strings.TrimSpace(s))
}

// MatchList ports match(List<AnalyzedSentence>) bug-for-bug with Java control flow.
func (r *PunctuationMarkAtParagraphEnd) MatchList(sentences []*languagetool.AnalyzedSentence) []*RuleMatch {
	var ruleMatches []*RuleMatch
	lastPara := -1
	pos := 0
	for n := 0; n < len(sentences); n++ {
		sentence := sentences[n]
		if r.isParagraphEnd(sentences, n) {
			tokens := sentence.GetTokensWithoutWhitespace()
			if len(tokens) > 2 {
				isFirstWord := (isParaWord(tokens[1]) && !isParaPunctuationMark(tokens[2])) ||
					(len(tokens) > 3 && isParaQuotationMark(tokens[1]) && isParaWord(tokens[2]) && !isParaPunctuationMark(tokens[3]))
				ignoreSentence := false
				if n == 1 && isParaNumeric(sentences[0].GetText()) {
					ignoreSentence = true
				}
				if n > 0 && isParaNumeric(sentences[n-1].GetText()) {
					ignoreSentence = true
				}
				// paragraphs with fewer than two sentences excluded (headlines, listings)
				if n-lastPara > 1 && isFirstWord && !ignoreSentence {
					lastNWToken := len(tokens) - 1
					for lastNWToken > 0 && tokens[lastNWToken].IsLinebreak() {
						lastNWToken--
					}
					// e.g. "find it at: http://example.com" should not be an error
					if len(tokens) >= 2 &&
						strings.EqualFold(tokens[len(tokens)-2].GetToken(), ":") &&
						tokenizers.IsURL(tokens[len(tokens)-1].GetToken()) {
						// Java: lastPara=n; pos += getText().length(); continue (skip normal pos += corrected)
						lastPara = n
						pos += utf16Len(sentence.GetText())
						continue
					}
					lastToken := tokens[lastNWToken].GetToken()
					low := strings.ToLower(lastToken)
					// Java: length > MAX && http || ftp  (&& binds tighter than ||)
					// Java continues without lastPara/pos update when long URL or ftp.
					if (utf16Len(lastToken) > maxURLLength && strings.HasPrefix(low, "http")) ||
						strings.HasPrefix(low, "ftp") {
						continue
					}
					if isParaWord(tokens[lastNWToken]) ||
						(isParaQuotationMark(tokens[lastNWToken]) && lastNWToken > 0 && isParaWord(tokens[lastNWToken-1])) {
						fromPos := pos + tokens[lastNWToken].GetStartPos()
						toPos := pos + tokens[lastNWToken].GetEndPos()
						msg := "Add a punctuation mark at paragraph end"
						if r.Messages != nil {
							if m := r.Messages["punctuation_mark_paragraph_end_msg"]; m != "" {
								msg = m
							}
						}
						rm := NewRuleMatch(r, sentence, fromPos, toPos, msg)
						var reps []string
						tok := tokens[lastNWToken].GetToken()
						for _, mark := range paraEndPunctMarks {
							reps = append(reps, tok+mark)
						}
						rm.SuggestedReplacements = reps
						ruleMatches = append(ruleMatches, rm)
					}
				}
			}
			lastPara = n
		}
		pos += sentence.GetCorrectedTextLength()
	}
	return ruleMatches
}
