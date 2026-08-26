package rules

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// PunctuationMarkAtParagraphEnd ports org.languagetool.rules.PunctuationMarkAtParagraphEnd.
// Java: Categories.PUNCTUATION, ITSIssueType.Grammar; Tag.picky;
// default ctor (messages, lang) → defaultActive true (not default-off).
// Language is represented by SingleLineBreaksMarksPara for Tools.isParagraphEnd
// (Demo/SRX SimpleSentenceTokenizer default false → need \n\n).
type PunctuationMarkAtParagraphEnd struct {
	Messages map[string]string
	// SingleLineBreaksMarksPara ports Language.getSentenceTokenizer().singleLineBreaksMarksPara()
	// as used by Tools.isParagraphEnd.
	SingleLineBreaksMarksPara bool
	Category                  *Category
	IssueType                 ITSIssueType
	DefaultOff                bool
	// Tags ports Rule.tags (Java setTags(Collections.singletonList(Tag.picky))).
	Tags []Tag
}

// Java PUNCTUATION_MARKS / QUOTATION_MARKS / P_NUMERIC / MAX_URL_LENGTH.
var (
	paraEndPunctMarks = []string{".", "!", "?", ":", ",", ";"}
	paraEndQuoteMarks = []string{"„", "»", "«", "\"", "”", "″", "’", "‚", "‘", "›", "‹", "′", "'"}
	paraEndNumericRE  = regexp.MustCompile(`^[0-9.]+$`)
)

const maxURLLength = 30 // Do not add punctuation after long URLs

// NewPunctuationMarkAtParagraphEnd ports
// PunctuationMarkAtParagraphEnd(ResourceBundle messages, Language lang)
// → this(messages, lang, true).
func NewPunctuationMarkAtParagraphEnd(messages map[string]string) *PunctuationMarkAtParagraphEnd {
	return NewPunctuationMarkAtParagraphEndDefaultActive(messages, true)
}

// NewPunctuationMarkAtParagraphEndDefaultActive ports
// PunctuationMarkAtParagraphEnd(ResourceBundle messages, Language lang, boolean defaultActive).
func NewPunctuationMarkAtParagraphEndDefaultActive(messages map[string]string, defaultActive bool) *PunctuationMarkAtParagraphEnd {
	r := &PunctuationMarkAtParagraphEnd{
		Messages:  messages,
		Category:  CatPunctuation.GetCategory(messages),
		IssueType: ITSGrammar,
		Tags:      []Tag{TagPicky},
	}
	if !defaultActive {
		r.DefaultOff = true
	}
	return r
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

// GetTags ports Rule.getTags (Java Tag.picky unless cleared).
func (r *PunctuationMarkAtParagraphEnd) GetTags() []Tag {
	if r == nil || len(r.Tags) == 0 {
		return nil
	}
	return append([]Tag(nil), r.Tags...)
}

// SetTags ports Rule.setTags (empty/nil clears).
func (r *PunctuationMarkAtParagraphEnd) SetTags(tags []Tag) {
	if r == nil {
		return
	}
	if len(tags) == 0 {
		r.Tags = nil
		return
	}
	r.Tags = append([]Tag(nil), tags...)
}

// HasTag ports Rule.hasTag.
func (r *PunctuationMarkAtParagraphEnd) HasTag(tag Tag) bool {
	if r == nil {
		return false
	}
	for _, t := range r.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// MinToCheckParagraph ports minToCheckParagraph (Java returns 0).
func (r *PunctuationMarkAtParagraphEnd) MinToCheckParagraph() int { return 0 }

func (r *PunctuationMarkAtParagraphEnd) isParagraphEnd(sentences []*languagetool.AnalyzedSentence, nTest int) bool {
	// Java: Tools.isParagraphEnd(sentences, n, lang)
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

// isParaWord ports isWord: Character.isLetter(token.charAt(0)).
func isParaWord(tk *languagetool.AnalyzedTokenReadings) bool {
	tok := tk.GetToken()
	if tok == "" {
		// Java charAt(0) would throw; empty is not a word in practice.
		return false
	}
	for _, r := range tok {
		return unicode.IsLetter(r)
	}
	return false
}

// isParaNumeric ports isNumeric: P_NUMERIC.matcher(s.trim()).matches() with [0-9.]+.
func isParaNumeric(s string) bool {
	return paraEndNumericRE.MatchString(tools.JavaStringTrim(s))
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
				// ignore sentences like "2.2.2. This is an item" (two sentences, first only numbers)
				ignoreSentence := false
				if n == 1 && isParaNumeric(sentences[0].GetText()) {
					ignoreSentence = true
				}
				if n > 0 && isParaNumeric(sentences[n-1].GetText()) {
					ignoreSentence = true
				}
				// paragraphs containing less than two sentences (e.g. headlines, listings) are excluded
				if n-lastPara > 1 && isFirstWord && !ignoreSentence {
					lastNWToken := len(tokens) - 1
					for lastNWToken > 0 && tokens[lastNWToken].IsLinebreak() {
						lastNWToken--
					}
					// e.g. "find it at: http://example.com" should not be an error
					if len(tokens) >= 2 &&
						strings.EqualFold(tokens[len(tokens)-2].GetToken(), ":") &&
						tokenizers.IsURL(tokens[len(tokens)-1].GetToken()) {
						// Java: lastPara=n; pos += getText().length(); continue
						// (skips normal pos += getCorrectedTextLength())
						lastPara = n
						pos += utf16Len(sentence.GetText())
						continue
					}
					lastToken := tokens[lastNWToken].GetToken()
					low := strings.ToLower(lastToken)
					// Java: length > MAX && http || ftp  (&& binds tighter than ||)
					// continue without lastPara/pos update (Java quirk).
					if (utf16Len(lastToken) > maxURLLength && strings.HasPrefix(low, "http")) ||
						strings.HasPrefix(low, "ftp") {
						continue
					}
					if isParaWord(tokens[lastNWToken]) ||
						(isParaQuotationMark(tokens[lastNWToken]) && lastNWToken > 0 && isParaWord(tokens[lastNWToken-1])) {
						fromPos := pos + tokens[lastNWToken].GetStartPos()
						toPos := pos + tokens[lastNWToken].GetEndPos()
						msg := "Please add a punctuation mark at the end of paragraph."
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
