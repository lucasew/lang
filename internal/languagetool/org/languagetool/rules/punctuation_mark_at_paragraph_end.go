package rules

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// PunctuationMarkAtParagraphEnd ports org.languagetool.rules.PunctuationMarkAtParagraphEnd.
type PunctuationMarkAtParagraphEnd struct {
	Messages map[string]string
	// SingleLineBreaksMarksPara matches Demo/SRX default false → need \n\n
	SingleLineBreaksMarksPara bool
}

var (
	paraEndPunctMarks = []string{".", "!", "?", ":", ",", ";"}
	paraEndQuoteMarks = []string{"„", "»", "«", "\"", "”", "″", "’", "‚", "‘", "›", "‹", "′", "'"}
	paraEndNumericRE  = regexp.MustCompile(`^[0-9.]+$`)
)

const maxURLLength = 30

func NewPunctuationMarkAtParagraphEnd(messages map[string]string) *PunctuationMarkAtParagraphEnd {
	return &PunctuationMarkAtParagraphEnd{Messages: messages}
}

func (r *PunctuationMarkAtParagraphEnd) GetID() string { return "PUNCTUATION_PARAGRAPH_END" }

func (r *PunctuationMarkAtParagraphEnd) isParagraphEnd(sentences []*languagetool.AnalyzedSentence, nTest int) bool {
	if nTest >= len(sentences)-1 {
		return true
	}
	text := sentences[nTest].GetText()
	if r.SingleLineBreaksMarksPara {
		if strings.HasSuffix(text, "\n") || strings.HasSuffix(text, "\n\r") {
			return true
		}
	} else {
		if strings.HasSuffix(text, "\n\n") || strings.HasSuffix(text, "\n\r\n\r") || strings.HasSuffix(text, "\r\n\r\n") {
			return true
		}
	}
	next := sentences[nTest+1].GetText()
	if strings.HasPrefix(next, "\n") || strings.HasPrefix(next, "\r\n") {
		return true
	}
	return false
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

// MatchList ports match(List<AnalyzedSentence>).
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
					colonURL := len(tokens) >= 2 &&
						strings.EqualFold(tokens[len(tokens)-2].GetToken(), ":") &&
						tokenizers.IsURL(tokens[len(tokens)-1].GetToken())
					lastToken := tokens[lastNWToken].GetToken()
					low := strings.ToLower(lastToken)
					// Java: length > MAX && http || ftp  (&& binds tighter than ||)
					longOrFTP := (utf16Len(lastToken) > maxURLLength && strings.HasPrefix(low, "http")) ||
						strings.HasPrefix(low, "ftp")
					if !colonURL && !longOrFTP {
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
			}
			lastPara = n
		}
		pos += sentence.GetCorrectedTextLength()
	}
	return ruleMatches
}
