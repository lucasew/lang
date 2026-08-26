package rules

import (
	"fmt"
	"strconv"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// WhitespaceCheckFilter ports org.languagetool.rules.WhitespaceCheckFilter.
// Line map (Java → Go):
//   acceptRuleMatch → AcceptRuleMatch
//   getRequired("whitespaceChar"|"position") → map lookup; missing → panic "Missing key '…'"
//   Integer.parseInt(position) → strconv.Atoi (invalid → panic)
//   pos < 1 || pos > patternTokens.length → IllegalArgumentException message via panic
//   !patternTokens[pos-1].getWhitespaceBefore().equals(wsChar) → return match; else null
type WhitespaceCheckFilter struct{}

func NewWhitespaceCheckFilter() *WhitespaceCheckFilter {
	return &WhitespaceCheckFilter{}
}

// AcceptRuleMatch ports WhitespaceCheckFilter.acceptRuleMatch.
// Required args: whitespaceChar, position (1-based into patternTokens).
func (f *WhitespaceCheckFilter) AcceptRuleMatch(match *RuleMatch, arguments map[string]string, _ int,
	patternTokens []*languagetool.AnalyzedTokenReadings, _ []int) *RuleMatch {
	// Java RuleFilter.getRequired("whitespaceChar", arguments)
	wsChar, ok := arguments["whitespaceChar"]
	if !ok {
		panic("Missing key 'whitespaceChar'")
	}
	// Java RuleFilter.getRequired("position", arguments)
	posStr, ok := arguments["position"]
	if !ok {
		panic("Missing key 'position'")
	}
	// Java: Integer.parseInt(...) — NumberFormatException on bad input
	pos, err := strconv.Atoi(posStr)
	if err != nil {
		panic(err)
	}
	if pos < 1 || pos > len(patternTokens) {
		// Java: throw new IllegalArgumentException("Wrong position in WhitespaceCheckFilter: " + pos + ", must be 1 to " + patternTokens.length)
		panic(fmt.Sprintf("Wrong position in WhitespaceCheckFilter: %d, must be 1 to %d", pos, len(patternTokens)))
	}
	// Java: patternTokens[pos - 1].getWhitespaceBefore().equals(wsChar)
	// null element → NPE (do not invent soft empty-string path)
	tok := patternTokens[pos-1]
	if tok == nil {
		panic("NullPointerException")
	}
	if tok.GetWhitespaceBefore() != wsChar {
		return match
	}
	return nil
}
