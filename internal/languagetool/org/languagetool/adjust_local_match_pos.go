package languagetool

// AdjustLocalMatchPos ports JLanguageTool.adjustRuleMatchPos for LocalMatch
// (cycle-free twin of rules.AdjustRuleMatchPos).
//
// charCount/columnCount/lineCount are the sentence's StartOffset/StartColumn/StartLine
// from ComputeSentenceData. sentence is the sentence plain text. All indices UTF-16.
// annotatedOriginal maps document positions when non-nil (AnnotatedText original).
func AdjustLocalMatchPos(
	m LocalMatch,
	charCount, columnCount, lineCount int,
	sentence string,
	mapOriginal func(pos int, isToPos bool) int, // nil = identity
) LocalMatch {
	fromPos := m.FromPos + charCount
	toPos := m.ToPos + charCount
	if mapOriginal != nil {
		fromPos = mapOriginal(fromPos, false)
		// Java: getOriginalTextPositionFor(toPos - 1, true) + 1
		if toPos > 0 {
			toPos = mapOriginal(toPos-1, true) + 1
		}
	}
	out := m
	// keep sentence-relative span
	if out.FromPosSentence < 0 || out.ToPosSentence <= out.FromPosSentence {
		out.FromPosSentence = m.FromPos
		out.ToPosSentence = m.ToPos
	}
	out.FromPos = fromPos
	out.ToPos = toPos

	// pattern defaults to error span (Java ctor sets pattern = from/to when not specified)
	patFrom, patTo := m.PatternFromPos, m.PatternToPos
	if patTo <= patFrom {
		patFrom, patTo = m.FromPos, m.ToPos
	}
	out.PatternFromPos = patFrom + charCount
	out.PatternToPos = patTo + charCount
	if mapOriginal != nil {
		out.PatternFromPos = mapOriginal(out.PatternFromPos, false)
		if out.PatternToPos > 0 {
			out.PatternToPos = mapOriginal(out.PatternToPos-1, true) + 1
		}
	}

	fromRel := m.FromPos
	toRel := m.ToPos
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
	partToError := utf16Prefix(sentence, fromRel)
	partToEnd := utf16Prefix(sentence, toRel)

	lastNL := utf16LastIndexOf(partToError, '\n')
	var column, endColumn int
	if lastNL == -1 {
		column = utf16Len(partToError) + columnCount
	} else {
		column = utf16Len(partToError) - lastNL
	}
	lastNLEnd := utf16LastIndexOf(partToEnd, '\n')
	if lastNLEnd == -1 {
		endColumn = utf16Len(partToEnd) + columnCount
	} else {
		endColumn = utf16Len(partToEnd) - lastNLEnd
	}
	out.Line = lineCount + CountLineBreaks(partToError)
	out.EndLine = lineCount + CountLineBreaks(partToEnd)
	out.Column = column
	out.EndColumn = endColumn
	if out.SentenceText == "" {
		out.SentenceText = sentence
	}
	return out
}
