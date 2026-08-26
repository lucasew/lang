package rules

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// UnderlineSpacesFilter ports org.languagetool.rules.UnderlineSpacesFilter.
// Expands match offsets to include adjacent whitespace before and/or after the marker.
type UnderlineSpacesFilter struct{}

func NewUnderlineSpacesFilter() *UnderlineSpacesFilter {
	return &UnderlineSpacesFilter{}
}

// AcceptRuleMatch ports UnderlineSpacesFilter.acceptRuleMatch.
// Required arg: underlineSpaces = before | after | both (other values leave offsets unchanged).
// Positions are Java String UTF-16 code units (sentence.length / substring).
func (f *UnderlineSpacesFilter) AcceptRuleMatch(match *RuleMatch, arguments map[string]string, _ int,
	_ []*languagetool.AnalyzedTokenReadings, _ []int) *RuleMatch {
	// Java RuleFilter.getRequired("underlineSpaces", arguments)
	underlineSpaces, ok := arguments["underlineSpaces"]
	if !ok {
		panic("Missing key 'underlineSpaces'")
	}
	// Java: match.getSentence().getText() — NPE if match or sentence is null
	// (AnalyzedSentence.GetText soft-returns "" on nil receiver; force NPE twin).
	if match == nil || match.Sentence == nil {
		panic("NullPointerException")
	}
	sentence := match.Sentence.GetText()
	if underlineSpaces == "before" || underlineSpaces == "both" {
		// Java: match.getFromPos() - 1 >= 0 && StringTools.isWhitespace(sentence.substring(from-1, from))
		if match.GetFromPos()-1 >= 0 &&
			tools.IsWhitespace(utf16Slice(sentence, match.GetFromPos()-1, match.GetFromPos())) {
			match.SetOffsetPosition(match.GetFromPos()-1, match.GetToPos())
		}
	}
	if underlineSpaces == "after" || underlineSpaces == "both" {
		// Java: match.getToPos() + 1 < sentence.length()
		// (stricter than substring bounds; trailing single whitespace is not expanded — Java quirk)
		if match.GetToPos()+1 < utf16Len(sentence) &&
			tools.IsWhitespace(utf16Slice(sentence, match.GetToPos(), match.GetToPos()+1)) {
			match.SetOffsetPosition(match.GetFromPos(), match.GetToPos()+1)
		}
	}
	return match
}
