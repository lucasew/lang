package rules

import (
	"fmt"
	"strconv"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// WhitespaceCheckFilter ports org.languagetool.rules.WhitespaceCheckFilter.
// Keeps the match when whitespace before the token at 1-based position
// is not equal to the expected whitespaceChar.
type WhitespaceCheckFilter struct{}

func NewWhitespaceCheckFilter() *WhitespaceCheckFilter {
	return &WhitespaceCheckFilter{}
}

// Accept returns true when the match should be kept (whitespace differs).
// position is 1-based into whitespaceBefore (same length as pattern tokens).
func (f *WhitespaceCheckFilter) Accept(whitespaceBefore []string, position int, whitespaceChar string) (keep bool, err string) {
	if position < 1 || position > len(whitespaceBefore) {
		return false, "Wrong position in WhitespaceCheckFilter"
	}
	return whitespaceBefore[position-1] != whitespaceChar, ""
}

// AcceptRuleMatch ports WhitespaceCheckFilter.acceptRuleMatch.
// Args: whitespaceChar (required), position (required, 1-based).
func (f *WhitespaceCheckFilter) AcceptRuleMatch(match *RuleMatch, arguments map[string]string, _ int,
	patternTokens []*languagetool.AnalyzedTokenReadings, _ []int) *RuleMatch {
	if f == nil || match == nil {
		return nil
	}
	wsChar, ok := arguments["whitespaceChar"]
	if !ok {
		panic("Missing key 'whitespaceChar'")
	}
	posStr, ok := arguments["position"]
	if !ok {
		panic("Missing key 'position'")
	}
	pos, err := strconv.Atoi(posStr)
	if err != nil {
		panic(err)
	}
	if pos < 1 || pos > len(patternTokens) {
		panic(fmt.Sprintf("Wrong position in WhitespaceCheckFilter: %d, must be 1 to %d", pos, len(patternTokens)))
	}
	tok := patternTokens[pos-1]
	got := ""
	if tok != nil {
		got = tok.GetWhitespaceBefore()
	}
	if got != wsChar {
		return match
	}
	return nil
}
