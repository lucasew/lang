package rules

import (
	"regexp"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// quoteSymbolLocator ports GenericUnpairedQuotesRule.SymbolLocator.
type quoteSymbolLocator struct {
	symbol   string
	startPos int
	sentence *languagetool.AnalyzedSentence
}

// GenericUnpairedQuotesRule ports org.languagetool.rules.GenericUnpairedQuotesRule.
// Java: TextLevelRule, Categories.PUNCTUATION, ITSIssueType.Typographical.
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
	// Optional language overrides (Java protected isNotBeginning/EndingApostrophe).
	// Nil → default GenericUnpairedQuotesRule logic.
	IsNotBeginningApostropheFn func(tokens []*languagetool.AnalyzedTokenReadings, i int) bool
	IsNotEndingApostropheFn    func(tokens []*languagetool.AnalyzedTokenReadings, i int) bool
}

// Java GenericUnpairedQuotesRule static patterns — Matcher.matches() = full string.
var (
	// POSSIBLE_APOSTROPHE: [‘’']
	possibleApostrophe = regexp.MustCompile(`^[‘’']$`)
	// INCH_PATTERN: .*\d".* with DOTALL
	inchPattern = regexp.MustCompile(`(?s).*\d".*`)
	// PUNCTUATION: [\p{Punct}…–—&&[^\"'_]]
	// Go: punct/ellipsis/dashes excluding " ' _
	quotePunctuation = regexp.MustCompile(`^[\p{P}…–—]$`)
	// PUNCT_MARKS: [\?\.!,]
	quotePunctMarks = regexp.MustCompile(`^[?.!,]$`)
)

// quotePunctuationOK ports PUNCTUATION.matcher(token).matches() excluding " ' _.
func quotePunctuationOK(s string) bool {
	if s == "\"" || s == "'" || s == "_" {
		return false
	}
	return quotePunctuation.MatchString(s)
}

// NewGenericUnpairedQuotesRule ports
// GenericUnpairedQuotesRule(messages, startSymbols, endSymbols).
func NewGenericUnpairedQuotesRule(messages map[string]string, start, end []string) *GenericUnpairedQuotesRule {
	return NewGenericUnpairedQuotesRuleWithID(nil, messages, start, end)
}

// NewGenericUnpairedQuotesRuleDefault ports
// GenericUnpairedQuotesRule(messages) with “” "" ‘’ ”.
func NewGenericUnpairedQuotesRuleDefault(messages map[string]string) *GenericUnpairedQuotesRule {
	return NewGenericUnpairedQuotesRule(messages,
		[]string{"“", "\"", "‘", "'"},
		[]string{"”", "\"", "’", "'"},
	)
}

// NewGenericUnpairedQuotesRuleWithID ports
// GenericUnpairedQuotesRule(ruleId, messages, startSymbols, endSymbols).
// ruleId nil/empty → "UNPAIRED_QUOTES".
func NewGenericUnpairedQuotesRuleWithID(ruleID *string, messages map[string]string, start, end []string) *GenericUnpairedQuotesRule {
	if len(start) != len(end) {
		panic("Different number of start and end symbols")
	}
	id := "UNPAIRED_QUOTES"
	if ruleID != nil && *ruleID != "" {
		id = *ruleID
	}
	return &GenericUnpairedQuotesRule{
		Messages:     messages,
		StartSymbols: start,
		EndSymbols:   end,
		ruleID:       id,
		Category:     CatPunctuation.GetCategory(messages),
		IssueType:    ITSTypographical,
	}
}

func (r *GenericUnpairedQuotesRule) GetID() string { return r.ruleID }

func (r *GenericUnpairedQuotesRule) SetRuleID(id string) { r.ruleID = id }

// GetDescription ports getDescription (desc_unpaired_quotes).
func (r *GenericUnpairedQuotesRule) GetDescription() string {
	if r != nil && r.Messages != nil {
		if m, ok := r.Messages["desc_unpaired_quotes"]; ok {
			return m
		}
	}
	return "Unpaired quotation marks"
}

// MinToCheckParagraph ports minToCheckParagraph (Java returns -1).
func (r *GenericUnpairedQuotesRule) MinToCheckParagraph() int { return -1 }

// EstimateContextForSureMatch ports TextLevelRule (Java always -1).
func (r *GenericUnpairedQuotesRule) EstimateContextForSureMatch() int { return -1 }

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

// isStartSymbolBefore ports isStartSymbolbefore.
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

// isNotOpenSymbol ports isNotOpenSymbol.
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

// isNotQuote ports isNotQuote.
func (r *GenericUnpairedQuotesRule) isNotQuote(tokens []*languagetool.AnalyzedTokenReadings, i, j int) bool {
	if (tokens[i-1].IsSentenceStart() || tokens[i].IsWhitespaceBefore()) &&
		(i >= len(tokens)-1 || tokens[i+1].IsWhitespaceBefore()) {
		return true
	}
	if r.EndSymbols[j] == r.StartSymbols[j] {
		if i < len(tokens)-1 &&
			!tokens[i].IsWhitespaceBefore() &&
			!tokens[i+1].IsWhitespaceBefore() &&
			quotePunctuationOK(tokens[i-1].GetToken()) &&
			tokens[i+1].GetToken() != "." &&
			quotePunctuationOK(tokens[i+1].GetToken()) {
			return true
		}
	}
	return false
}

// isOpeningQuote ports isOpeningQuote.
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
							quotePunctuationOK(tokens[i-1].GetToken())) ||
							hasSuffixDash(tokens[i-1].GetToken()))) ||
					r.isStartSymbolBefore(tokens, i)
			}
			return true
		}
	}
	return false
}

// hasSuffixDash ports tokens[i-1].getToken().endsWith("-").
func hasSuffixDash(s string) bool {
	return len(s) > 0 && s[len(s)-1] == '-'
}

// isClosingQuote ports isClosingQuote.
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

// isInchQuote ports isInchQuote.
func (r *GenericUnpairedQuotesRule) isInchQuote(text string) bool {
	return inchPattern.MatchString(text)
}

// isNotBeginningApostrophe ports isNotBeginningApostrophe.
func (r *GenericUnpairedQuotesRule) isNotBeginningApostrophe(tokens []*languagetool.AnalyzedTokenReadings, i int) bool {
	if r != nil && r.IsNotBeginningApostropheFn != nil {
		return r.IsNotBeginningApostropheFn(tokens, i)
	}
	return !possibleApostrophe.MatchString(tokens[i].GetToken()) ||
		i >= len(tokens)-1 || tokens[i+1].IsNonWord() || tokens[i+1].IsWhitespaceBefore()
}

// isNotEndingApostrophe ports isNotEndingApostrophe.
func (r *GenericUnpairedQuotesRule) isNotEndingApostrophe(tokens []*languagetool.AnalyzedTokenReadings, i int) bool {
	if r != nil && r.IsNotEndingApostropheFn != nil {
		return r.IsNotEndingApostropheFn(tokens, i)
	}
	return !possibleApostrophe.MatchString(tokens[i].GetToken()) ||
		tokens[i].IsWhitespaceBefore() ||
		tokens[i-1].IsNonWord()
}

// indexOfOpeningQuote ports indexOfOpeningQuote.
func (r *GenericUnpairedQuotesRule) indexOfOpeningQuote(openingQuotes []quoteSymbolLocator, symbol string) int {
	for i, oq := range openingQuotes {
		if symbol == oq.symbol {
			return i
		}
	}
	return -1
}

// addMatch ports addMatch.
func (r *GenericUnpairedQuotesRule) addMatch(opening quoteSymbolLocator, ruleMatches *[]*RuleMatch) {
	other := r.findCorrespondingSymbol(opening.symbol)
	// Java: MessageFormat.format(messages.getString("unpaired_brackets"), findCorrespondingSymbol(...))
	pattern := "Unpaired symbol: ''{0}'' seems to be missing"
	if r != nil && r.Messages != nil {
		if m, ok := r.Messages["unpaired_brackets"]; ok && m != "" {
			pattern = m
		}
	}
	msg := tools.I18n(pattern, other)
	// Java: symbol.length() (UTF-16 code units)
	symLen := utf16Len(opening.symbol)
	*ruleMatches = append(*ruleMatches, NewRuleMatch(r, opening.sentence, opening.startPos, opening.startPos+symLen, msg))
}

// removeAllOpenInnerQuotes ports removeAllOpenInnerQuotes.
func (r *GenericUnpairedQuotesRule) removeAllOpenInnerQuotes(index int, openingQuotes *[]quoteSymbolLocator, ruleMatches *[]*RuleMatch) {
	for i := len(*openingQuotes) - 1; i > index; i-- {
		r.addMatch((*openingQuotes)[i], ruleMatches)
		*openingQuotes = append((*openingQuotes)[:i], (*openingQuotes)[i+1:]...)
	}
}

// findCorrespondingSymbol ports findCorrespondingSymbol.
func (r *GenericUnpairedQuotesRule) findCorrespondingSymbol(symbol string) string {
	if idx := indexOf(r.StartSymbols, symbol); idx >= 0 {
		return r.EndSymbols[idx]
	}
	idx := indexOf(r.EndSymbols, symbol)
	return r.StartSymbols[idx]
}

// FindCorrespondingSymbol is the exported twin of protected findCorrespondingSymbol.
func (r *GenericUnpairedQuotesRule) FindCorrespondingSymbol(symbol string) string {
	return r.findCorrespondingSymbol(symbol)
}
