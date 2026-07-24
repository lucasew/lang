package rules

import (
	"regexp"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// SymbolType ports GenericUnpairedBracketsRule.Symbol.Type.
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
// Index is the token index in the sentence (fillSymbolStack) or the
// ruleMatches list index (createMatch / ruleMatchStack), matching Java reuse.
type SymbolLocator struct {
	Symbol       BracketSymbol
	Index        int
	StartPos     int
	Sentence     *languagetool.AnalyzedSentence
	SentenceIdx  int
	MatchListIdx int // ruleMatchStack: same role as Java SymbolLocator.index
}

// GenericUnpairedBracketsRule ports org.languagetool.rules.GenericUnpairedBracketsRule.
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

	// Optional subclass hooks (Java protected overrides). Nil → base behavior.
	// IsNoExceptionFn ports isNoException.
	IsNoExceptionFn func(token string, tokens []*languagetool.AnalyzedTokenReadings, i, j int, precSpace, follSpace bool, symbolStack *UnsyncStack[SymbolLocator]) bool
	// PreventMatchFn ports preventMatch.
	PreventMatchFn func(sentence *languagetool.AnalyzedSentence) bool
	// GetSuggestionsFn ports getSuggestions.
	GetSuggestionsFn func(text func() string, startPos, endPos int, symbol BracketSymbol, otherSymbol string) []string
	// GetMessageFn ports getMessage; empty/nil → MessageFormat(unpaired_brackets).
	GetMessageFn func() string
}

// Java GenericUnpairedBracketsRule static patterns (Matcher.matches() = full string).
var (
	// NUMERALS_EN — Java ends with $; matches() anchors both ends.
	defaultNumerals = regexp.MustCompile(`(?i)^(\d{1,2}?[a-z']*|M*(D?C{0,3}|C[DM])(L?X{0,3}|X[LC])(V?I{0,3}|I[VX]))$`)
	// PUNCTUATION: [\p{Punct}…–—]
	gubrPunctuation = regexp.MustCompile(`^[\p{P}…–—]$`)
	// PUNCTUATION_NO_DOT first alt: [ldmnstLDMNST]'
	gubrApostropheElision = regexp.MustCompile(`^[ldmnstLDMNST]'$`)
	// URL exception: tokens[i-1].matches("https?://.+")
	gubrURLToken = regexp.MustCompile(`^https?://.+$`)
	// createMatch enumeration-label contexts (String.matches full)
	gubrEnumAfterNL = regexp.MustCompile(`^\n[a-zA-Z]\)$`)
	gubrEnumAtStart = regexp.MustCompile(`^[a-zA-Z]\)$`)
)

// NewGenericUnpairedBracketsRule ports
// GenericUnpairedBracketsRule(messages, startSymbols, endSymbols).
func NewGenericUnpairedBracketsRule(messages map[string]string, start, end []string) *GenericUnpairedBracketsRule {
	return NewGenericUnpairedBracketsRuleWithID(nil, messages, start, end, nil)
}

// NewGenericUnpairedBracketsRuleWithNumerals ports the 4-arg Java ctor (custom numeral pattern).
// When numerals is nil, uses NUMERALS_EN.
func NewGenericUnpairedBracketsRuleWithNumerals(messages map[string]string, start, end []string, numerals *regexp.Regexp) *GenericUnpairedBracketsRule {
	return NewGenericUnpairedBracketsRuleWithID(nil, messages, start, end, numerals)
}

// NewGenericUnpairedBracketsRuleWithID ports
// GenericUnpairedBracketsRule(ruleId, messages, startSymbols, endSymbols[, numerals]).
// ruleId nil → "UNPAIRED_BRACKETS".
func NewGenericUnpairedBracketsRuleWithID(ruleID *string, messages map[string]string, start, end []string, numerals *regexp.Regexp) *GenericUnpairedBracketsRule {
	if len(start) != len(end) {
		panic("Different number of start and end symbols")
	}
	uniqueMap := uniqueMapInit(end)
	if numerals == nil {
		numerals = defaultNumerals
	}
	id := "UNPAIRED_BRACKETS"
	if ruleID != nil && *ruleID != "" {
		id = *ruleID
	}
	return &GenericUnpairedBracketsRule{
		Messages:     messages,
		Category:     CatPunctuation.GetCategory(messages),
		IssueType:    ITSTypographical,
		StartSymbols: start,
		EndSymbols:   end,
		uniqueMap:    uniqueMap,
		numerals:     numerals,
		ruleID:       id,
	}
}

// NewGenericUnpairedBracketsRuleDefault ports
// GenericUnpairedBracketsRule(messages) with [] () {} "" ''.
func NewGenericUnpairedBracketsRuleDefault(messages map[string]string) *GenericUnpairedBracketsRule {
	return NewGenericUnpairedBracketsRule(messages,
		[]string{"[", "(", "{", "\"", "'"},
		[]string{"]", ")", "}", "\"", "'"},
	)
}

func uniqueMapInit(endSymbols []string) map[string]bool {
	uniqueMap := map[string]bool{}
	for _, es := range endSymbols {
		found := 0
		for _, e2 := range endSymbols {
			if e2 == es {
				found++
			}
		}
		uniqueMap[es] = found == 1
	}
	return uniqueMap
}

func (r *GenericUnpairedBracketsRule) GetID() string { return r.ruleID }

func (r *GenericUnpairedBracketsRule) SetRuleID(id string) { r.ruleID = id }

// GetDescription ports getDescription (desc_unpaired_brackets).
func (r *GenericUnpairedBracketsRule) GetDescription() string {
	if r != nil && r.Messages != nil {
		if m, ok := r.Messages["desc_unpaired_brackets"]; ok {
			return m
		}
	}
	return "Unpaired braces, brackets, quotation marks and similar symbols"
}

// MinToCheckParagraph ports minToCheckParagraph (Java returns -1).
func (r *GenericUnpairedBracketsRule) MinToCheckParagraph() int { return -1 }

// EstimateContextForSureMatch ports TextLevelRule (Java always -1).
func (r *GenericUnpairedBracketsRule) EstimateContextForSureMatch() int { return -1 }

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

// MatchList ports GenericUnpairedBracketsRule.match(List<AnalyzedSentence>).
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

	// if the stack is odd and symmetric match only the symbol in the middle, e.g. ({"})
	// Java compares stack.get(i) vs stack.get(ssSize - 1) — not ssSize-1-i (bug-for-bug).
	isSymmetric := false
	ssSize := symbolStack.Len()
	if ssSize > 2 && ssSize%2 == 1 {
		isSymmetric = true
		for i := 0; i < ssSize/2; i++ {
			if indexOf(r.StartSymbols, symbolStack.At(i).Symbol.Symbol) !=
				indexOf(r.EndSymbols, symbolStack.At(ssSize-1).Symbol.Symbol) {
				isSymmetric = false
				break
			}
		}
	}

	// Supplier.memoize full text
	var (
		fullOnce sync.Once
		fullText string
	)
	lazyFullText := func() string {
		fullOnce.Do(func() {
			var b strings.Builder
			for _, a := range sentences {
				b.WriteString(a.GetText())
			}
			fullText = b.String()
		})
		return fullText
	}

	if isSymmetric {
		loc := symbolStack.At(ssSize / 2)
		rm := r.createMatch(&ruleMatches, ruleMatchStack, loc.StartPos, loc.Symbol, loc.Sentence, loc.SentenceIdx, lazyFullText)
		if rm != nil {
			ruleMatches = append(ruleMatches, rm)
		}
	} else {
		for _, sLoc := range symbolStack.Data() {
			rm := r.createMatch(&ruleMatches, ruleMatchStack, sLoc.StartPos, sLoc.Symbol, sLoc.Sentence, sLoc.SentenceIdx, lazyFullText)
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
	noException := r.isNoException(token, tokens, i, j, precededByWhitespace, isSpecialCase, symbolStack)

	if noException && precededByWhitespace && token == r.StartSymbols[j] {
		symbolStack.Push(SymbolLocator{
			Symbol: BracketSymbol{Symbol: r.StartSymbols[j], Type: SymbolOpening},
			Index:  i, StartPos: startPos, Sentence: sentence, SentenceIdx: sentenceIdx,
		})
		return true
	} else if noException && (isSpecialCase || tokens[i].IsSentenceEnd()) && token == r.EndSymbols[j] {
		// Java: empty-body skip for numeral enumeration closers like "1)", "1.)"
		if r.isNumeralEnumerationCloser(tokens, i, j, symbolStack) {
			// fall through — do not push/pop (Java empty if body)
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
		if r.isEndSymbolUnique(r.EndSymbols[j]) {
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
	// Java: numerals.matches(...) && !(!empty && peek=="(")
	openParenOnStack := !symbolStack.Empty() && symbolStack.Peek().Symbol.Symbol == "("
	// form: <numeral> . )  e.g. "1.)" / "XII.)"
	if i > 2 && tokens[i-1] != nil && tokens[i-2] != nil && tokens[i-3] != nil {
		if tokens[i-1].GetToken() == "." &&
			r.numerals.MatchString(tokens[i-2].GetToken()) &&
			(tokens[i-3].HasPosTag("SENT_START") || tokens[i-2].IsWhitespaceBefore()) &&
			!openParenOnStack {
			return true
		}
	}
	// form: <numeral> )  e.g. "1)", "1a)", "iv)"
	if i > 1 && tokens[i-1] != nil && r.numerals.MatchString(tokens[i-1].GetToken()) && !openParenOnStack {
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

// isNoException ports GenericUnpairedBracketsRule.isNoException.
// true means "no exception" → treat as a real bracket symbol.
func (r *GenericUnpairedBracketsRule) isNoException(token string, tokens []*languagetool.AnalyzedTokenReadings, i, j int, precSpace, follSpace bool, symbolStack *UnsyncStack[SymbolLocator]) bool {
	if r != nil && r.IsNoExceptionFn != nil {
		return r.IsNoExceptionFn(token, tokens, i, j, precSpace, follSpace, symbolStack)
	}
	return r.defaultIsNoException(token, tokens, i, j, precSpace, follSpace, symbolStack)
}

// DefaultIsNoException is the base-class isNoException body (for subclass super calls).
func (r *GenericUnpairedBracketsRule) DefaultIsNoException(token string, tokens []*languagetool.AnalyzedTokenReadings, i, j int, precSpace, follSpace bool, symbolStack *UnsyncStack[SymbolLocator]) bool {
	return r.defaultIsNoException(token, tokens, i, j, precSpace, follSpace, symbolStack)
}

func (r *GenericUnpairedBracketsRule) defaultIsNoException(token string, tokens []*languagetool.AnalyzedTokenReadings, i, j int, precSpace, follSpace bool, symbolStack *UnsyncStack[SymbolLocator]) bool {
	_ = token
	_ = j
	_ = precSpace
	_ = follSpace
	_ = symbolStack
	tokenStr := tokens[i].GetToken()
	// Java: tokens[i-1].matches("https?://.+") && contains("(")
	if i > 0 {
		prev := tokens[i-1].GetToken()
		if gubrURLToken.MatchString(prev) && strings.Contains(prev, "(") {
			return false
		}
	}
	if i >= 2 {
		prevPrev := tokens[i-2].GetToken()
		prev := tokens[i-1].GetToken()
		// Smiley ":-)" and ":-("
		if prevPrev == ":" && prev == "-" && (tokenStr == ")" || tokenStr == "(") {
			return false
		}
		// Smiley ";-)" and ";-("
		if prevPrev == ";" && prev == "-" && (tokenStr == ")" || tokenStr == "(") {
			return false
		}
	}
	if i >= 1 {
		prev := tokens[i-1].GetToken()
		// Smiley ":)" and ":("
		if prev == ":" && !tokens[i].IsWhitespaceBefore() && (tokenStr == ")" || tokenStr == "(") {
			return false
		}
		// Smiley ";)" and ";("
		if prev == ";" && !tokens[i].IsWhitespaceBefore() && (tokenStr == ")" || tokenStr == "(") {
			return false
		}
	}
	return true
}

func (r *GenericUnpairedBracketsRule) isEndSymbolUnique(str string) bool {
	return r.uniqueMap[str]
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
		Symbol: symbol, Index: len(*ruleMatches), MatchListIdx: len(*ruleMatches),
		StartPos: startPos, Sentence: sentence, SentenceIdx: sentenceIdx,
	})
	other := r.FindCorrespondingSymbol(symbol)
	msg := r.messageFor(other)
	fullText := lazyFullText()
	// Java uses String.length() (UTF-16 code units)
	symLen := utf16Len(symbol.Symbol)
	fullLen := utf16Len(fullText)
	if startPos+symLen < fullLen {
		if startPos >= 2 && startPos+symLen < fullLen {
			context := utf16Substring(fullText, startPos-2, startPos+symLen)
			if gubrEnumAfterNL.MatchString(context) { // prevent error for "b) foo item"
				return nil
			}
		} else if startPos >= 1 {
			context := utf16Substring(fullText, startPos-1, startPos+symLen)
			if gubrEnumAtStart.MatchString(context) { // prevent error for "a) foo item" at text start
				return nil
			}
		}
	}
	if r.preventMatch(sentence) {
		return nil
	}
	match := NewRuleMatch(r, sentence, startPos, startPos+symLen, msg)
	if repl := r.getSuggestions(lazyFullText, startPos, startPos+symLen, symbol, other); repl != nil {
		match.SetSuggestedReplacements(repl)
	}
	return match
}

func (r *GenericUnpairedBracketsRule) messageFor(otherSymbol string) string {
	if r != nil && r.GetMessageFn != nil {
		if m := r.GetMessageFn(); m != "" {
			return m
		}
	}
	// Java: MessageFormat.format(messages.getString("unpaired_brackets"), otherSymbol)
	pattern := "Unpaired symbol: ''{0}'' seems to be missing"
	if r != nil && r.Messages != nil {
		if m, ok := r.Messages["unpaired_brackets"]; ok && m != "" {
			pattern = m
		}
	}
	return tools.I18n(pattern, otherSymbol)
}

func (r *GenericUnpairedBracketsRule) preventMatch(sentence *languagetool.AnalyzedSentence) bool {
	if r != nil && r.PreventMatchFn != nil {
		return r.PreventMatchFn(sentence)
	}
	return false
}

func (r *GenericUnpairedBracketsRule) getSuggestions(text func() string, startPos, endPos int, symbol BracketSymbol, otherSymbol string) []string {
	if r != nil && r.GetSuggestionsFn != nil {
		return r.GetSuggestionsFn(text, startPos, endPos, symbol, otherSymbol)
	}
	return nil
}

// FindCorrespondingSymbol ports findCorrespondingSymbol.
func (r *GenericUnpairedBracketsRule) FindCorrespondingSymbol(symbol BracketSymbol) string {
	if idx := indexOf(r.StartSymbols, symbol.Symbol); idx >= 0 {
		return r.EndSymbols[idx]
	}
	idx := indexOf(r.EndSymbols, symbol.Symbol)
	if idx < 0 {
		return ""
	}
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
	return gubrPunctuation.MatchString(s)
}

func isPunctuationNoDot(s string) bool {
	// Java: [ldmnstLDMNST]'|[–—\p{Punct}&&[^.]]
	if gubrApostropheElision.MatchString(s) {
		return true
	}
	if s == "." {
		return false
	}
	return gubrPunctuation.MatchString(s)
}
