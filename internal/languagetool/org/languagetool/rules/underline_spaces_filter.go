package rules

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"

// UnderlineSpacesFilter ports org.languagetool.rules.UnderlineSpacesFilter.
// Expands match offsets to include adjacent whitespace before/after/both.
type UnderlineSpacesFilter struct{}

func NewUnderlineSpacesFilter() *UnderlineSpacesFilter {
	return &UnderlineSpacesFilter{}
}

// Expand returns possibly extended fromPos/toPos for the match.
// underlineSpaces is "before", "after", or "both".
func (f *UnderlineSpacesFilter) Expand(sentence string, fromPos, toPos int, underlineSpaces string) (int, int) {
	if underlineSpaces == "before" || underlineSpaces == "both" {
		if fromPos-1 >= 0 && tools.IsWhitespace(sentence[fromPos-1:fromPos]) {
			fromPos--
		}
	}
	if underlineSpaces == "after" || underlineSpaces == "both" {
		if toPos+1 <= len(sentence) && toPos < len(sentence) && tools.IsWhitespace(sentence[toPos:toPos+1]) {
			toPos++
		}
	}
	return fromPos, toPos
}
