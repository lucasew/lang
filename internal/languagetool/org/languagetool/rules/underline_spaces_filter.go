package rules

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// UnderlineSpacesFilter ports org.languagetool.rules.UnderlineSpacesFilter.
// Expands match offsets to include adjacent whitespace before/after/both.
type UnderlineSpacesFilter struct{}

func NewUnderlineSpacesFilter() *UnderlineSpacesFilter {
	return &UnderlineSpacesFilter{}
}

// Expand returns possibly extended fromPos/toPos for the match.
// underlineSpaces is "before", "after", or "both".
// Positions are the same code-unit convention as the match (Java String UTF-16).
func (f *UnderlineSpacesFilter) Expand(sentence string, fromPos, toPos int, underlineSpaces string) (int, int) {
	if underlineSpaces == "before" || underlineSpaces == "both" {
		if fromPos-1 >= 0 {
			ch := utf16Slice(sentence, fromPos-1, fromPos)
			if tools.IsWhitespace(ch) {
				fromPos--
			}
		}
	}
	if underlineSpaces == "after" || underlineSpaces == "both" {
		if toPos+1 <= utf16Len(sentence) {
			ch := utf16Slice(sentence, toPos, toPos+1)
			if tools.IsWhitespace(ch) {
				toPos++
			}
		}
	}
	return fromPos, toPos
}

// AcceptRuleMatch ports UnderlineSpacesFilter.acceptRuleMatch.
// Args: underlineSpaces = before | after | both (required).
func (f *UnderlineSpacesFilter) AcceptRuleMatch(match *RuleMatch, arguments map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *RuleMatch {
	if f == nil || match == nil {
		return nil
	}
	mode, ok := arguments["underlineSpaces"]
	if !ok {
		panic("Missing key 'underlineSpaces'")
	}
	var sentence string
	if match.Sentence != nil {
		sentence = match.Sentence.GetText()
	}
	from, to := f.Expand(sentence, match.GetFromPos(), match.GetToPos(), mode)
	match.SetOffsetPosition(from, to)
	return match
}
