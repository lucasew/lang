package rules

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/markup"
)

// AdjustRuleMatchPos ports JLanguageTool.adjustRuleMatchPos:
// shift sentence-local match offsets to document positions, map through
// AnnotatedText when present, and set line/column from sentence + base counts.
//
// All character indices are Java UTF-16 code units.
func AdjustRuleMatchPos(
	match *RuleMatch,
	charCount, columnCount, lineCount int,
	sentence string,
	annotatedText *markup.AnnotatedText,
) *RuleMatch {
	if match == nil {
		return nil
	}
	fromPos := match.GetFromPos() + charCount
	toPos := match.GetToPos() + charCount
	if annotatedText != nil {
		fromPos = annotatedText.GetOriginalTextPositionFor(fromPos, false)
		// Java: getOriginalTextPositionFor(toPos - 1, true) + 1
		toPos = annotatedText.GetOriginalTextPositionFor(toPos-1, true) + 1
	}
	thisMatch := CloneRuleMatch(match)
	thisMatch.SetOffsetPosition(fromPos, toPos)
	// keep positions with respect to sentence start
	thisMatch.SetSentencePosition(match.GetFromPos(), match.GetToPos())

	startPos := match.GetPatternFromPos() + charCount
	endPos := match.GetPatternToPos() + charCount
	thisMatch.SetPatternPosition(startPos, endPos)

	// Java uses String.substring / lastIndexOf with UTF-16 indices.
	fromRel := match.GetFromPos()
	toRel := match.GetToPos()
	if fromRel < 0 {
		fromRel = 0
	}
	if toRel < 0 {
		toRel = 0
	}
	sentLen := utf16Len(sentence)
	if fromRel > sentLen {
		fromRel = sentLen
	}
	if toRel > sentLen {
		toRel = sentLen
	}
	sentencePartToError := utf16Substring(sentence, 0, fromRel)
	sentencePartToEndOfError := utf16Substring(sentence, 0, toRel)

	lastLineBreakPos := utf16LastIndexOfRune(sentencePartToError, '\n')
	var column, endColumn int
	if lastLineBreakPos == -1 {
		column = utf16Len(sentencePartToError) + columnCount
	} else {
		column = utf16Len(sentencePartToError) - lastLineBreakPos
	}
	lastLineBreakPosInError := utf16LastIndexOfRune(sentencePartToEndOfError, '\n')
	if lastLineBreakPosInError == -1 {
		endColumn = utf16Len(sentencePartToEndOfError) + columnCount
	} else {
		endColumn = utf16Len(sentencePartToEndOfError) - lastLineBreakPosInError
	}
	lineBreaksToError := languagetool.CountLineBreaks(sentencePartToError)
	lineBreaksToEndOfError := languagetool.CountLineBreaks(sentencePartToEndOfError)
	thisMatch.SetLine(lineCount + lineBreaksToError)
	thisMatch.SetEndLine(lineCount + lineBreaksToEndOfError)
	thisMatch.SetColumn(column)
	thisMatch.SetEndColumn(endColumn)
	return thisMatch
}

// utf16LastIndexOfRune returns UTF-16 index of last r, or -1.
func utf16LastIndexOfRune(s string, r rune) int {
	idx := -1
	u := 0
	for _, ch := range s {
		if ch == r {
			idx = u
		}
		if ch >= 0x10000 {
			u += 2
		} else {
			u++
		}
	}
	return idx
}
