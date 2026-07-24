package rules

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// SymbolType for unpaired brackets stack.
type SymbolType int

const (
	SymbolOpening SymbolType = iota
	SymbolClosing
)

// BracketSymbol ports GenericUnpairedBracketsRule.Symbol.
type BracketSymbol struct {
	Symbol string
	Type   SymbolType
}

// SymbolLocator ports org.languagetool.rules.SymbolLocator.
type SymbolLocator struct {
	Symbol       BracketSymbol
	Index        int
	StartPos     int
	Sentence     *languagetool.AnalyzedSentence
	SentenceIdx  int
	MatchListIdx int // for ruleMatchStack
}

// GenericUnpairedBracketsRule ports org.languagetool.rules.GenericUnpairedBracketsRule.
// Java: PUNCTUATION, Typographical.
type GenericUnpairedBracketsRule struct {
	Messages     map[string]string
	StartSymbols []string
	EndSymbols   []string
	Category     *Category
	IssueType    ITSIssueType
	// URL ports Rule.url (Java setUrl).
	URL string
	// incorrectExamples / correctExamples port Rule.addExamplePair.
	incorrectExamples []IncorrectExample
	correctExamples   []CorrectExample
	uniqueMap         map[string]bool
	numerals          *regexp.Regexp
	ruleID            string
}

// defaultNumerals is Java GenericUnpairedBracketsRule.NUMERALS_EN.
// Java Pattern used with Matcher.matches() (full string); Go anchors both alternatives.
var defaultNumerals = regexp.MustCompile(`(?i)^(\d{1,2}?[a-z']*|M*(D?C{0,3}|C[DM])(L?X{0,3}|X[LC])(V?I{0,3}|I[VX]))$`)

func NewGenericUnpairedBracketsRule(messages map[string]string, start, end []string) *GenericUnpairedBracketsRule {
	return NewGenericUnpairedBracketsRuleWithNumerals(messages, start, end, nil)
}

// NewGenericUnpairedBracketsRuleWithNumerals ports the 4-arg Java ctor (custom numeral pattern).
// When numerals is nil, uses the default English-style numeral pattern.
func NewGenericUnpairedBracketsRuleWithNumerals(messages map[string]string, start, end []string, numerals *regexp.Regexp) *GenericUnpairedBracketsRule {
	if len(start) != len(end) {
		panic("start/end symbol count mismatch")
	}
	uniqueMap := map[string]bool{}
	for _, es := range end {
		found := 0
		for _, e2 := range end {
			if e2 == es {
				found++
			}
		}
		uniqueMap[es] = found == 1
	}
	if numerals == nil {
		numerals = defaultNumerals
	}
	return &GenericUnpairedBracketsRule{
		Messages:     messages,
		Category:     CatPunctuation.GetCategory(messages),
		IssueType:    ITSTypographical,
		StartSymbols: start,
		EndSymbols:   end,
		uniqueMap:    uniqueMap,
		numerals:     numerals,
		ruleID:       "UNPAIRED_BRACKETS",
	}
}

func (r *GenericUnpairedBracketsRule) GetID() string { return r.ruleID }

func (r *GenericUnpairedBracketsRule) SetRuleID(id string) { r.ruleID = id }

func (r *GenericUnpairedBracketsRule) GetCategory() *Category {
	if r == nil {
		return nil
	}
	return r.Category
}

func (r *GenericUnpairedBracketsRule) GetLocQualityIssueType() ITSIssueType {
	if r == nil || r.IssueType == "" {
		return ITSTypographical
	}
	return r.IssueType
}

// GetURL ports Rule.getUrl.
func (r *GenericUnpairedBracketsRule) GetURL() string {
	if r == nil {
		return ""
	}
	return r.URL
}

// SetURL ports Rule.setUrl.
func (r *GenericUnpairedBracketsRule) SetURL(u string) {
	if r != nil {
		r.URL = u
	}
}

// AddExamplePair ports Rule.addExamplePair.
func (r *GenericUnpairedBracketsRule) AddExamplePair(incorrect IncorrectExample, correct CorrectExample) {
	if r == nil {
		return
	}
	appendExamplePair(&r.incorrectExamples, &r.correctExamples, incorrect, correct)
}

// GetIncorrectExamples ports Rule.getIncorrectExamples.
func (r *GenericUnpairedBracketsRule) GetIncorrectExamples() []IncorrectExample {
	if r == nil || len(r.incorrectExamples) == 0 {
		return nil
	}
	out := make([]IncorrectExample, len(r.incorrectExamples))
	copy(out, r.incorrectExamples)
	return out
}

// GetCorrectExamples ports Rule.getCorrectExamples.
func (r *GenericUnpairedBracketsRule) GetCorrectExamples() []CorrectExample {
	if r == nil || len(r.correctExamples) == 0 {
		return nil
	}
	out := make([]CorrectExample, len(r.correctExamples))
	copy(out, r.correctExamples)
	return out
}

func (r *GenericUnpairedBracketsRule) MatchList(sentences []*languagetool.AnalyzedSentence) []*RuleMatch {
	symbolStack := NewUnsyncStack[SymbolLocator]()
	ruleMatchStack := NewUnsyncStack[SymbolLocator]()
	var ruleMatches []*RuleMatch
	startPosBase := 0
	sentenceIdx := 0
	for _, sentence := range sentences {
		tokens := sentence.GetTokensWithoutWhitespace()
		for i := 1; i < len(tokens); i++ {
			for j := 0; j < len(r.StartSymbols); j++ {
				if r.fillSymbolStack(startPosBase, tokens, i, j, symbolStack, sentence, sentenceIdx) {
					break
				}
			}
		}
		startPosBase += sentence.GetCorrectedTextLength()
		sentenceIdx++
	}

	isSymmetric := false
	ssSize := symbolStack.Len()
	if ssSize > 2 && ssSize%2 == 1 {
		isSymmetric = true
		for i := 0; i < ssSize/2; i++ {
			if indexOf(r.StartSymbols, symbolStack.At(i).Symbol.Symbol) !=
				indexOf(r.EndSymbols, symbolStack.At(ssSize-1-i).Symbol.Symbol) {
				isSymmetric = false
				break
			}
		}
	}

	fullText := func() string {
		var b strings.Builder
		for _, a := range sentences {
			b.WriteString(a.GetText())
		}
		return b.String()
	}

	if isSymmetric {
		loc := symbolStack.At(ssSize / 2)
		rm := r.createMatch(&ruleMatches, ruleMatchStack, loc.StartPos, loc.Symbol, loc.Sentence, loc.SentenceIdx, fullText)
		if rm != nil {
			ruleMatches = append(ruleMatches, rm)
		}
	} else {
		for _, sLoc := range symbolStack.Data() {
			rm := r.createMatch(&ruleMatches, ruleMatchStack, sLoc.StartPos, sLoc.Symbol, sLoc.Sentence, sLoc.SentenceIdx, fullText)
			if rm != nil && (sLoc.Symbol.Type == SymbolClosing ||
				endsLikeRealSentence(sLoc.Sentence.GetText()) ||
				len(sentences)-1 > sLoc.SentenceIdx) {
				ruleMatches = append(ruleMatches, rm)
			}
		}
	}
	return ruleMatches
}

func endsLikeRealSentence(r string) bool {
	// Java: r.trim() then endsWith . ? !
	s := tools.JavaStringTrim(r)
	return strings.HasSuffix(s, ".") || strings.HasSuffix(s, "?") || strings.HasSuffix(s, "!")
}

func (r *GenericUnpairedBracketsRule) fillSymbolStack(startPosBase int, tokens []*languagetool.AnalyzedTokenReadings, i, j int, symbolStack *UnsyncStack[SymbolLocator], sentence *languagetool.AnalyzedSentence, sentenceIdx int) bool {
	token := tokens[i].GetToken()
	startPos := startPosBase + tokens[i].GetStartPos()
	if token != r.StartSymbols[j] && token != r.EndSymbols[j] {
		return false
	}
	precededByWhitespace := r.getPrecededByWhitespace(tokens, i, j)
	isSpecialCase := r.getSpecialCase(tokens, i, j)
	noException := r.isNoException(token, tokens, i, j, precededByWhitespace, isSpecialCase)

	if noException && precededByWhitespace && token == r.StartSymbols[j] {
		symbolStack.Push(SymbolLocator{
			Symbol: BracketSymbol{Symbol: r.StartSymbols[j], Type: SymbolOpening},
			Index:  i, StartPos: startPos, Sentence: sentence, SentenceIdx: sentenceIdx,
		})
		return true
	} else if noException && (isSpecialCase || tokens[i].IsSentenceEnd()) && token == r.EndSymbols[j] {
		// Java: skip numeral enumeration closers like "1)", "1a)", "iv)", "1.)"
		// when not pairing an open "(".
		if r.isNumeralEnumerationCloser(tokens, i, j, symbolStack) {
			return false
		}
		if symbolStack.Empty() {
			symbolStack.Push(SymbolLocator{
				Symbol: BracketSymbol{Symbol: r.EndSymbols[j], Type: SymbolClosing},
				Index:  i, StartPos: startPos, Sentence: sentence, SentenceIdx: sentenceIdx,
			})
			return true
		}
		if symbolStack.Peek().Symbol.Symbol == r.StartSymbols[j] {
			symbolStack.Pop()
			return true
		}
		if r.uniqueMap[r.EndSymbols[j]] {
			symbolStack.Push(SymbolLocator{
				Symbol: BracketSymbol{Symbol: r.EndSymbols[j], Type: SymbolClosing},
				Index:  i, StartPos: startPos, Sentence: sentence, SentenceIdx: sentenceIdx,
			})
			return true
		}
		if j == len(r.EndSymbols)-1 {
			symbolStack.Push(SymbolLocator{
				Symbol: BracketSymbol{Symbol: r.EndSymbols[j], Type: SymbolClosing},
				Index:  i, StartPos: startPos, Sentence: sentence, SentenceIdx: sentenceIdx,
			})
			return true
		}
	}
	return false
}

// isNumeralEnumerationCloser ports the empty-body skip in Java fillSymbolStack for
// endSymbols ")" after numerals (Chapter 1). / section 1a). / XII.)).
func (r *GenericUnpairedBracketsRule) isNumeralEnumerationCloser(tokens []*languagetool.AnalyzedTokenReadings, i, j int, symbolStack *UnsyncStack[SymbolLocator]) bool {
	if r == nil || r.EndSymbols[j] != ")" || r.numerals == nil {
		return false
	}
	// open paren on stack → this ) is a real closer, not an enum label
	if !symbolStack.Empty() && symbolStack.Peek().Symbol.Symbol == "(" {
		return false
	}
	// form: <numeral> . )  e.g. "1.)" / "XII.)"
	if i > 2 && tokens[i-1] != nil && tokens[i-2] != nil && tokens[i-3] != nil {
		if tokens[i-1].GetToken() == "." &&
			r.numerals.MatchString(tokens[i-2].GetToken()) &&
			(tokens[i-3].IsSentenceStart() || tokens[i-2].IsWhitespaceBefore()) {
			return true
		}
	}
	// form: <numeral> )  e.g. "1)", "1a)", "iv)"
	if i > 1 && tokens[i-1] != nil && r.numerals.MatchString(tokens[i-1].GetToken()) {
		return true
	}
	return false
}

func (r *GenericUnpairedBracketsRule) getPrecededByWhitespace(tokens []*languagetool.AnalyzedTokenReadings, i, j int) bool {
	precededByWhitespace := true
	if r.StartSymbols[j] == r.EndSymbols[j] {
		precededByWhitespace = tokens[i-1].IsSentenceStart() ||
			tokens[i].IsWhitespaceBefore() ||
			isPunctuationNoDot(tokens[i-1].GetToken()) ||
			containsStr(r.StartSymbols, tokens[i-1].GetToken())
	}
	return precededByWhitespace
}

func (r *GenericUnpairedBracketsRule) getSpecialCase(tokens []*languagetool.AnalyzedTokenReadings, i, j int) bool {
	isException := true
	if i < len(tokens)-1 && r.StartSymbols[j] == r.EndSymbols[j] {
		isException = tokens[i+1].IsWhitespaceBefore() ||
			isPunctuation(tokens[i+1].GetToken()) ||
			containsStr(r.EndSymbols, tokens[i+1].GetToken()) ||
			(i >= 1 && strings.HasSuffix(tokens[i-1].GetToken(), "-")) ||
			strings.HasPrefix(tokens[i+1].GetToken(), "-") ||
			tokens[i+1].GetToken() == "s"
	}
	return isException
}

func (r *GenericUnpairedBracketsRule) isNoException(token string, tokens []*languagetool.AnalyzedTokenReadings, i, j int, precSpace, follSpace bool) bool {
	tokenStr := tokens[i].GetToken()
	// Java: URL token containing '(' — brackets inside/after URL are not unpaired
	// (tokens[i-1].matches("https?://.+") && contains("(")).
	if i > 0 {
		prev := tokens[i-1].GetToken()
		if (strings.HasPrefix(prev, "http://") || strings.HasPrefix(prev, "https://")) &&
			strings.Contains(prev, "(") {
			return false
		}
	}
	if i >= 2 {
		prevPrev := tokens[i-2].GetToken()
		prev := tokens[i-1].GetToken()
		if prevPrev == ":" && prev == "-" && (tokenStr == ")" || tokenStr == "(") {
			return false
		}
		if prevPrev == ";" && prev == "-" && (tokenStr == ")" || tokenStr == "(") {
			return false
		}
	}
	if i >= 1 {
		prev := tokens[i-1].GetToken()
		if prev == ":" && !tokens[i].IsWhitespaceBefore() && (tokenStr == ")" || tokenStr == "(") {
			return false
		}
		if prev == ";" && !tokens[i].IsWhitespaceBefore() && (tokenStr == ")" || tokenStr == "(") {
			return false
		}
	}
	return true
}

func (r *GenericUnpairedBracketsRule) createMatch(ruleMatches *[]*RuleMatch, ruleMatchStack *UnsyncStack[SymbolLocator], startPos int, symbol BracketSymbol, sentence *languagetool.AnalyzedSentence, sentenceIdx int, lazyFullText func() string) *RuleMatch {
	if !ruleMatchStack.Empty() {
		index := indexOf(r.EndSymbols, symbol.Symbol)
		if index >= 0 {
			rLoc := ruleMatchStack.Peek()
			if rLoc.Symbol.Symbol == r.StartSymbols[index] {
				if len(*ruleMatches) > rLoc.MatchListIdx {
					// remove paired match
					idx := rLoc.MatchListIdx
					*ruleMatches = append((*ruleMatches)[:idx], (*ruleMatches)[idx+1:]...)
					ruleMatchStack.Pop()
					return nil
				}
			}
		}
	}
	ruleMatchStack.Push(SymbolLocator{
		Symbol: symbol, MatchListIdx: len(*ruleMatches), StartPos: startPos,
		Sentence: sentence, SentenceIdx: sentenceIdx,
	})
	other := r.findCorresponding(symbol)
	msg := fmt.Sprintf("Unpaired bracket: expected %s", other)
	if r.Messages != nil {
		if m, ok := r.Messages["unpaired_brackets"]; ok {
			msg = fmt.Sprintf(m, other)
			if !strings.Contains(m, "%") && !strings.Contains(m, "{") {
				msg = m + " " + other
			}
		}
	}
	fullText := lazyFullText()
	symLen := utf16Len(symbol.Symbol)
	if startPos+symLen < utf16Len(fullText) {
		if startPos >= 2 {
			context := utf16Substring(fullText, startPos-2, startPos+symLen)
			if matched, _ := regexp.MatchString(`\n[a-zA-Z]\)`, context); matched {
				return nil
			}
		} else if startPos >= 1 {
			context := utf16Substring(fullText, startPos-1, startPos+symLen)
			if matched, _ := regexp.MatchString(`^[a-zA-Z]\)$`, context); matched {
				return nil
			}
		}
	}
	return NewRuleMatch(r, sentence, startPos, startPos+symLen, msg)
}

func (r *GenericUnpairedBracketsRule) findCorresponding(symbol BracketSymbol) string {
	if idx := indexOf(r.StartSymbols, symbol.Symbol); idx >= 0 {
		return r.EndSymbols[idx]
	}
	idx := indexOf(r.EndSymbols, symbol.Symbol)
	return r.StartSymbols[idx]
}

func indexOf(ss []string, s string) int {
	for i, x := range ss {
		if x == s {
			return i
		}
	}
	return -1
}

func containsStr(ss []string, s string) bool { return indexOf(ss, s) >= 0 }

func isPunctuation(s string) bool {
	ok, _ := regexp.MatchString(`^[\p{P}…–—]$`, s)
	return ok
}

func isPunctuationNoDot(s string) bool {
	// [ldmnstLDMNST]'|–—punct without .
	if matched, _ := regexp.MatchString(`^[ldmnstLDMNST]'$`, s); matched {
		return true
	}
	ok, _ := regexp.MatchString(`^[–—\p{P}]$`, s)
	if !ok {
		return false
	}
	return s != "."
}
