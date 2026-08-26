package rules

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// AddCommasFilter ports org.languagetool.rules.AddCommasFilter.
type AddCommasFilter struct{}

func NewAddCommasFilter() *AddCommasFilter {
	return &AddCommasFilter{}
}

// Java OPENING_QUOTES.matcher(token).matches() — whole-token match.
var openingQuotesRE = regexp.MustCompile(`^[«“"‘'„¿¡]$`)

// AcceptRuleMatch ports AddCommasFilter.acceptRuleMatch (Java).
func (f *AddCommasFilter) AcceptRuleMatch(match *RuleMatch, arguments map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *RuleMatch {
	if f == nil || match == nil || match.Sentence == nil {
		return nil
	}
	bSuggestSemicolon := strings.EqualFold(arguments["suggestSemicolon"], "true")
	tokens := match.Sentence.GetTokensWithoutWhitespace()
	if len(tokens) == 0 {
		return nil
	}
	postagFrom := 1
	for postagFrom < len(tokens) && tokens[postagFrom] != nil && tokens[postagFrom].GetStartPos() < match.GetFromPos() {
		postagFrom++
	}
	postagTo := postagFrom
	for postagTo < len(tokens) && tokens[postagTo] != nil && tokens[postagTo].GetEndPos() < match.GetToPos() {
		postagTo++
	}
	if postagFrom >= len(tokens) || tokens[postagFrom] == nil {
		return nil
	}
	if postagTo >= len(tokens) || tokens[postagTo] == nil {
		return nil
	}

	beforeOK := postagFrom == 1 ||
		(postagFrom > 0 && tokens[postagFrom-1] != nil && tools.IsPunctuationOrSymbol(tokens[postagFrom-1].GetToken())) ||
		tools.IsCapitalizedWord(tokens[postagFrom].GetToken())
	afterOK := !(postagTo+1 > len(tokens)-1) &&
		tokens[postagTo+1] != nil &&
		tools.IsPunctuationOrSymbol(tokens[postagTo+1].GetToken()) &&
		!(tokens[postagTo+1].IsWhitespaceBefore() && openingQuotesRE.MatchString(tokens[postagTo+1].GetToken()))
	if beforeOK && afterOK {
		return nil
	}

	// Positions are UTF-16 code units (Java String indices).
	matched := utf16Substring(match.Sentence.GetText(), match.GetFromPos(), match.GetToPos())
	msg := match.GetMessage()
	short := match.ShortMessage

	if bSuggestSemicolon && postagFrom > 0 && tokens[postagFrom-1] != nil &&
		tokens[postagFrom-1].GetToken() == "," && !afterOK {
		newMatch := NewRuleMatch(match.GetRule(), match.Sentence,
			tokens[postagFrom-1].GetStartPos(), tokens[postagTo].GetEndPos(), msg)
		newMatch.ShortMessage = short
		newMatch.SetSuggestedReplacements([]string{
			"; " + matched + ",",
			", " + matched + ",",
		})
		return newMatch
	}
	if beforeOK && !afterOK {
		newMatch := NewRuleMatch(match.GetRule(), match.Sentence,
			tokens[postagTo].GetStartPos(), match.GetToPos(), msg)
		newMatch.ShortMessage = short
		newMatch.SetSuggestedReplacement(tokens[postagTo].GetToken() + ",")
		return newMatch
	}
	if !beforeOK && afterOK {
		startPos := tokens[postagFrom].GetStartPos()
		if tokens[postagFrom].IsWhitespaceBefore() {
			startPos--
		}
		newMatch := NewRuleMatch(match.GetRule(), match.Sentence,
			startPos, tokens[postagFrom].GetEndPos(), msg)
		newMatch.ShortMessage = short
		newMatch.SetSuggestedReplacement(", " + tokens[postagFrom].GetToken())
		return newMatch
	}
	// !beforeOK && !afterOK
	startPos := tokens[postagFrom].GetStartPos()
	if tokens[postagFrom].IsWhitespaceBefore() {
		startPos--
	}
	newMatch := NewRuleMatch(match.GetRule(), match.Sentence,
		startPos, tokens[postagTo].GetEndPos(), msg)
	newMatch.ShortMessage = short
	newMatch.SetSuggestedReplacement(", " + matched + ",")
	return newMatch
}

// CommaContext + Suggest kept for unit tests of neighborhood logic.
type CommaContext struct {
	MatchedText              string
	FirstToken, LastToken    string
	TokenBefore              string
	TokenAfter               string
	AfterHasWhitespaceBefore bool
	SuggestSemicolon         bool
	MatchAtSentenceStart     bool
}

func (f *AddCommasFilter) Accept(ctx CommaContext) bool {
	beforeOK, afterOK := f.sidesOK(ctx)
	return !(beforeOK && afterOK)
}

func (f *AddCommasFilter) Suggest(ctx CommaContext) []string {
	beforeOK, afterOK := f.sidesOK(ctx)
	if beforeOK && afterOK {
		return nil
	}
	if ctx.SuggestSemicolon && ctx.TokenBefore == "," && !afterOK {
		return []string{
			"; " + ctx.MatchedText + ",",
			", " + ctx.MatchedText + ",",
		}
	}
	if beforeOK && !afterOK {
		return []string{ctx.LastToken + ","}
	}
	if !beforeOK && afterOK {
		return []string{", " + ctx.FirstToken}
	}
	return []string{", " + ctx.MatchedText + ","}
}

func (f *AddCommasFilter) sidesOK(ctx CommaContext) (beforeOK, afterOK bool) {
	beforeOK = ctx.MatchAtSentenceStart || ctx.TokenBefore == "" ||
		tools.IsPunctuationOrSymbol(ctx.TokenBefore) || tools.IsCapitalizedWord(ctx.FirstToken)
	afterOK = ctx.TokenAfter != "" && tools.IsPunctuationOrSymbol(ctx.TokenAfter) &&
		!(ctx.AfterHasWhitespaceBefore && openingQuotesRE.MatchString(ctx.TokenAfter))
	return beforeOK, afterOK
}
