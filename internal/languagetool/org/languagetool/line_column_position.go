package languagetool

import "strings"

// LineColumnPosition ports JLanguageTool.TextCheckCallable.LineColumnPosition
// (0-based line/column as stored on RuleMatch before CLI 1-based display).
type LineColumnPosition struct {
	Line   int
	Column int
}

// NewLineColumnPosition constructs a LineColumnPosition.
func NewLineColumnPosition(line, column int) LineColumnPosition {
	return LineColumnPosition{Line: line, Column: column}
}

// CountLineBreaks ports JLanguageTool.countLineBreaks.
func CountLineBreaks(s string) int {
	count := 0
	pos := -1
	for {
		next := strings.IndexByte(s[pos+1:], '\n')
		if next < 0 {
			break
		}
		pos = pos + 1 + next
		count++
	}
	return count
}

// ProcessColumnChange ports processColumnChange(columnCount, sentence)
// without the singleLineBreaksMarksPara special case (pass false → no --).
func ProcessColumnChange(columnCount int, sentence string) int {
	return ProcessColumnChangePara(columnCount, sentence, false)
}

// ProcessColumnChangePara includes singleLineBreaksMarksPara behavior.
func ProcessColumnChangePara(columnCount int, sentence string, singleLineBreaksMarksPara bool) int {
	lineBreakPos := strings.LastIndexByte(sentence, '\n')
	if lineBreakPos == -1 {
		return columnCount + len(sentence)
	}
	columnCount = len(sentence) - lineBreakPos
	if lineBreakPos == 0 && !singleLineBreaksMarksPara {
		columnCount--
	}
	return columnCount
}

// FindLineColumnInSentences ports TextCheckCallable.findLineColumn.
// sentences must be ordered by StartOffset; offset is a document position.
func FindLineColumnInSentences(sentences []SentenceData, offset int) LineColumnPosition {
	if len(sentences) == 0 {
		return NewLineColumnPosition(0, 0)
	}
	sentence := findSentenceContaining(sentences, offset)
	rel := offset - sentence.StartOffset
	if rel < 0 {
		rel = 0
	}
	if rel > len(sentence.Text) {
		rel = len(sentence.Text)
	}
	prefix := sentence.Text[:rel]
	return NewLineColumnPosition(
		sentence.StartLine+CountLineBreaks(prefix),
		ProcessColumnChange(sentence.StartColumn, prefix),
	)
}

func findSentenceContaining(sentences []SentenceData, offset int) SentenceData {
	low, high := 0, len(sentences)-1
	for low <= high {
		mid := (low + high) / 2
		sentence := sentences[mid]
		if sentence.StartOffset < offset {
			low = mid + 1
		} else if sentence.StartOffset > offset {
			high = mid - 1
		} else {
			return sentence
		}
	}
	if low-1 < 0 {
		return sentences[0]
	}
	return sentences[low-1]
}
