package rules

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// quoteSymbolLocator ports GenericUnpairedQuotesRule.SymbolLocator.
type quoteSymbolLocator struct {
	symbol   string
	startPos int
	sentence *languagetool.AnalyzedSentence
}

// GenericUnpairedQuotesRule ports org.languagetool.rules.GenericUnpairedQuotesRule.
// Java: PUNCTUATION, Typographical.
type GenericUnpairedQuotesRule struct {
	Messages     map[string]string
	StartSymbols []string
	EndSymbols   []string
	ruleID       string
	// Category ports Rule.category (Java PUNCTUATION).
	Category *Category
	// IssueType ports getLocQualityIssueType (Java Typographical).
	IssueType ITSIssueType
	// URL ports Rule.url (Java setUrl).
	URL string
	// incorrectExamples / correctExamples port Rule.addExamplePair.
	incorrectExamples []IncorrectExample
	correctExamples   []CorrectExample
	// Optional language overrides (Java subclass isNotBeginning/EndingApostrophe).
	// Nil → default GenericUnpairedQuotesRule logic.
	IsNotBeginningApostropheFn func(tokens []*languagetool.AnalyzedTokenReadings, i int) bool
	IsNotEndingApostropheFn    func(tokens []*languagetool.AnalyzedTokenReadings, i int) bool
}

var (
	possibleApostrophe = regexp.MustCompile(`[‘’']`)
	inchPattern        = regexp.MustCompile(`(?s).*\d".*`)
	// Java: [\p{Punct}…–—&&[^\"'_]] — punct including …–— excluding " ' _
	quotePunctuation = func(s string) bool {
		if s == "" || s == "\"" || s == "'" || s == "_" {
			return false
		}
		if s == "…" || s == "–" || s == "—" {
			return true
		}
		// unicode.IsPunct covers \p{Punct}
		for _, r := range s {
			if !unicode.IsPunct(r) {
				return false
			}
		}
		return true
	}
	quotePunctMarks = regexp.MustCompile(`^[?.!,]$`)
)

func NewGenericUnpairedQuotesRule(messages map[string]string, start, end []string) *GenericUnpairedQuotesRule {
	if len(start) != len(end) {
		panic("start/end symbol count mismatch")
	}
	return &GenericUnpairedQuotesRule{
		Messages:     messages,
		StartSymbols: start,
		EndSymbols:   end,
		ruleID:       "UNPAIRED_QUOTES",
		Category:     CatPunctuation.GetCategory(messages),
		IssueType:    ITSTypographical,
	}
}

func (r *GenericUnpairedQuotesRule) GetID() string { return r.ruleID }

func (r *GenericUnpairedQuotesRule) SetRuleID(id string) { r.ruleID = id }

// GetCategory ports Rule.getCategory.
func (r *GenericUnpairedQuotesRule) GetCategory() *Category {
	if r == nil {
		return nil
	}
	return r.Category
}

// GetLocQualityIssueType ports Rule.getLocQualityIssueType.
func (r *GenericUnpairedQuotesRule) GetLocQualityIssueType() ITSIssueType {
	if r == nil || r.IssueType == "" {
		return ITSTypographical
	}
	return r.IssueType
}

// GetURL ports Rule.getUrl.
func (r *GenericUnpairedQuotesRule) GetURL() string {
	if r == nil {
		return ""
	}
	return r.URL
}

// SetURL ports Rule.setUrl.
func (r *GenericUnpairedQuotesRule) SetURL(u string) {
	if r != nil {
		r.URL = u
	}
}

// AddExamplePair ports Rule.addExamplePair.
func (r *GenericUnpairedQuotesRule) AddExamplePair(incorrect IncorrectExample, correct CorrectExample) {
	if r == nil {
		return
	}
	appendExamplePair(&r.incorrectExamples, &r.correctExamples, incorrect, correct)
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *GenericUnpairedQuotesRule) GetIncorrectExamples() []IncorrectExample {
	if r == nil || len(r.incorrectExamples) == 0 {
		return nil
	}
	out := make([]IncorrectExample, len(r.incorrectExamples))
	copy(out, r.incorrectExamples)
	return out
}

// GetCorrectExamples ports Rule.getCorrectExamples.
func (r *GenericUnpairedQuotesRule) GetCorrectExamples() []CorrectExample {
	if r == nil || len(r.correctExamples) == 0 {
		return nil
	}
	out := make([]CorrectExample, len(r.correctExamples))
	copy(out, r.correctExamples)
	return out
}

// MatchList ports match(List<AnalyzedSentence>).
func (r *GenericUnpairedQuotesRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*RuleMatch {
	var openingQuotes []quoteSymbolLocator
	var ruleMatches []*RuleMatch
	var lastApostropheSymbol *string
	wasInch := false
	startPosBase := 0
	for _, sentence := range sentences {
		tokens := sentence.GetTokensWithoutWhitespace()
		for i := 1; i < len(tokens); i++ {
			if r.isOpeningQuote(tokens, i) {
				symbol := tokens[i].GetToken()
				if !r.isNotBeginningApostrophe(tokens, i) {
					s := symbol
					lastApostropheSymbol = &s
					continue
				}
				if symbol == "\"" {
					wasInch = false
				}
				if lastApostropheSymbol != nil && *lastApostropheSymbol == symbol {
					lastApostropheSymbol = nil
				}
				idx := r.indexOfOpeningQuote(openingQuotes, symbol)
				if idx >= 0 {
					r.removeAllOpenInnerQuotes(idx-1, &openingQuotes, &ruleMatches)
				}
				openingQuotes = append(openingQuotes, quoteSymbolLocator{
					symbol:   symbol,
					startPos: tokens[i].GetStartPos() + startPosBase,
					sentence: sentence,
				})
			} else if r.isClosingQuote(tokens, i, openingQuotes) {
				symbol := tokens[i].GetToken()
				if !r.isNotBeginningApostrophe(tokens, i) {
					s := symbol
					lastApostropheSymbol = &s
					continue
				}
				isInchSymb := symbol == "\""
				isInch := false
				if isInchSymb {
					isInch = r.isInchQuote(sentence.GetText())
				}
				startSymbol := r.findCorrespondingSymbol(symbol)
				idx := r.indexOfOpeningQuote(openingQuotes, startSymbol)
				if idx >= 0 {
					r.removeAllOpenInnerQuotes(idx, &openingQuotes, &ruleMatches)
					// remove opening at idx
					openingQuotes = append(openingQuotes[:idx], openingQuotes[idx+1:]...)
					if lastApostropheSymbol != nil && *lastApostropheSymbol == startSymbol {
						lastApostropheSymbol = nil
					}
					if isInch {
						wasInch = true
					}
				} else if r.isNotEndingApostrophe(tokens, i) {
					if !isInch && (!isInchSymb || !wasInch) {
						if lastApostropheSymbol == nil || *lastApostropheSymbol != symbol {
							r.addMatch(quoteSymbolLocator{
								symbol:   symbol,
								startPos: tokens[i].GetStartPos() + startPosBase,
								sentence: sentence,
							}, &ruleMatches)
						} else {
							lastApostropheSymbol = nil
						}
					} else {
						wasInch = false
					}
				}
			}
		}
		startPosBase += sentence.GetCorrectedTextLength()
	}
	r.removeAllOpenInnerQuotes(-1, &openingQuotes, &ruleMatches)
	return ruleMatches
}

func (r *GenericUnpairedQuotesRule) isStartSymbolBefore(tokens []*languagetool.AnalyzedTokenReadings, i int) bool {
	for j := i - 1; j > 0; j-- {
		if tokens[i].GetToken() != tokens[j].GetToken() && containsStr(r.StartSymbols, tokens[j].GetToken()) {
			if tokens[j-1].IsSentenceStart() || tokens[j].IsWhitespaceBefore() {
				return true
			}
		} else {
			return false
		}
	}
	return true
}

func (r *GenericUnpairedQuotesRule) isNotOpenSymbol(j int, openingQuotes []quoteSymbolLocator) bool {
	if r.EndSymbols[j] == r.StartSymbols[j] {
		for _, oq := range openingQuotes {
			if r.EndSymbols[j] == oq.symbol {
				return false
			}
		}
	}
	return true
}

func (r *GenericUnpairedQuotesRule) isNotQuote(tokens []*languagetool.AnalyzedTokenReadings, i, j int) bool {
	if (tokens[i-1].IsSentenceStart() || tokens[i].IsWhitespaceBefore()) &&
		(i >= len(tokens)-1 || tokens[i+1].IsWhitespaceBefore()) {
		return true
	}
	if r.EndSymbols[j] == r.StartSymbols[j] {
		if i < len(tokens)-1 &&
			!tokens[i].IsWhitespaceBefore() &&
			!tokens[i+1].IsWhitespaceBefore() &&
			quotePunctuation(tokens[i-1].GetToken()) &&
			tokens[i+1].GetToken() != "." &&
			quotePunctuation(tokens[i+1].GetToken()) {
			return true
		}
	}
	return false
}

func (r *GenericUnpairedQuotesRule) isOpeningQuote(tokens []*languagetool.AnalyzedTokenReadings, i int) bool {
	for j := 0; j < len(r.StartSymbols); j++ {
		if r.StartSymbols[j] == tokens[i].GetToken() {
			if r.isNotQuote(tokens, i, j) {
				return false
			}
			if containsStr(r.EndSymbols, r.StartSymbols[j]) {
				return tokens[i-1].IsSentenceStart() ||
					tokens[i].IsWhitespaceBefore() ||
					(i < len(tokens)-1 && !tokens[i+1].IsWhitespaceBefore() &&
						((!quotePunctMarks.MatchString(tokens[i+1].GetToken()) &&
							quotePunctuation(tokens[i-1].GetToken())) ||
							strings.HasSuffix(tokens[i-1].GetToken(), "-"))) ||
					r.isStartSymbolBefore(tokens, i)
			}
			return true
		}
	}
	return false
}

func (r *GenericUnpairedQuotesRule) isClosingQuote(tokens []*languagetool.AnalyzedTokenReadings, i int, openingQuotes []quoteSymbolLocator) bool {
	for j := 0; j < len(r.EndSymbols); j++ {
		if r.EndSymbols[j] == tokens[i].GetToken() {
			if r.isNotQuote(tokens, i, j) && r.isNotOpenSymbol(j, openingQuotes) {
				return false
			}
			return true
		}
	}
	return false
}

func (r *GenericUnpairedQuotesRule) isInchQuote(text string) bool {
	return inchPattern.MatchString(text)
}

func (r *GenericUnpairedQuotesRule) isNotBeginningApostrophe(tokens []*languagetool.AnalyzedTokenReadings, i int) bool {
	if r != nil && r.IsNotBeginningApostropheFn != nil {
		return r.IsNotBeginningApostropheFn(tokens, i)
	}
	return !possibleApostrophe.MatchString(tokens[i].GetToken()) ||
		i >= len(tokens)-1 || tokens[i+1].IsNonWord() || tokens[i+1].IsWhitespaceBefore()
}

func (r *GenericUnpairedQuotesRule) isNotEndingApostrophe(tokens []*languagetool.AnalyzedTokenReadings, i int) bool {
	if r != nil && r.IsNotEndingApostropheFn != nil {
		return r.IsNotEndingApostropheFn(tokens, i)
	}
	return !possibleApostrophe.MatchString(tokens[i].GetToken()) ||
		tokens[i].IsWhitespaceBefore() ||
		tokens[i-1].IsNonWord()
}

func (r *GenericUnpairedQuotesRule) indexOfOpeningQuote(openingQuotes []quoteSymbolLocator, symbol string) int {
	for i, oq := range openingQuotes {
		if symbol == oq.symbol {
			return i
		}
	}
	return -1
}

func (r *GenericUnpairedQuotesRule) addMatch(opening quoteSymbolLocator, ruleMatches *[]*RuleMatch) {
	other := r.findCorrespondingSymbol(opening.symbol)
	msg := fmt.Sprintf("Unpaired quotes: expected %s", other)
	if r.Messages != nil {
		if m, ok := r.Messages["unpaired_brackets"]; ok {
			if strings.Contains(m, "%s") || strings.Contains(m, "%v") {
				msg = fmt.Sprintf(m, other)
			} else if strings.Contains(m, "{0}") {
				msg = strings.ReplaceAll(m, "{0}", other)
			} else {
				msg = m + " " + other
			}
		}
	}
	symLen := utf16Len(opening.symbol)
	*ruleMatches = append(*ruleMatches, NewRuleMatch(r, opening.sentence, opening.startPos, opening.startPos+symLen, msg))
}

func (r *GenericUnpairedQuotesRule) removeAllOpenInnerQuotes(index int, openingQuotes *[]quoteSymbolLocator, ruleMatches *[]*RuleMatch) {
	for i := len(*openingQuotes) - 1; i > index; i-- {
		r.addMatch((*openingQuotes)[i], ruleMatches)
		*openingQuotes = append((*openingQuotes)[:i], (*openingQuotes)[i+1:]...)
	}
}

func (r *GenericUnpairedQuotesRule) findCorrespondingSymbol(symbol string) string {
	if idx := indexOf(r.StartSymbols, symbol); idx >= 0 {
		return r.EndSymbols[idx]
	}
	idx := indexOf(r.EndSymbols, symbol)
	return r.StartSymbols[idx]
}
