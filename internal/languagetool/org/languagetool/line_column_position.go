package languagetool

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

// CountLineBreaks ports JLanguageTool.countLineBreaks
// (Java String.indexOf('\n') over UTF-16 code units; \n is a single unit).
func CountLineBreaks(s string) int {
	// Counting rune/byte occurrences of '\n' matches UTF-16 for this ASCII char.
	count := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			count++
		}
	}
	return count
}

// ProcessColumnChange ports processColumnChange(columnCount, sentence)
// without the singleLineBreaksMarksPara special case (pass false → no --).
func ProcessColumnChange(columnCount int, sentence string) int {
	return ProcessColumnChangePara(columnCount, sentence, false)
}

// ProcessColumnChangePara ports JLanguageTool.processColumnChange with the
// language.getSentenceTokenizer().singleLineBreaksMarksPara() flag.
// All lengths / indices use Java String UTF-16 code units.
func ProcessColumnChangePara(columnCount int, sentence string, singleLineBreaksMarksPara bool) int {
	lineBreakPos := utf16LastIndexOf(sentence, '\n')
	if lineBreakPos == -1 {
		return columnCount + utf16Len(sentence)
	}
	columnCount = utf16Len(sentence) - lineBreakPos
	if lineBreakPos == 0 && !singleLineBreaksMarksPara {
		columnCount--
	}
	return columnCount
}

// FindLineColumnInSentences ports TextCheckCallable.findLineColumn.
// sentences must be ordered by StartOffset; offset is a document UTF-16 position.
func FindLineColumnInSentences(sentences []SentenceData, offset int) LineColumnPosition {
	if len(sentences) == 0 {
		return NewLineColumnPosition(0, 0)
	}
	sentence := findSentenceContaining(sentences, offset)
	rel := offset - sentence.StartOffset
	if rel < 0 {
		rel = 0
	}
	tlen := utf16Len(sentence.Text)
	if rel > tlen {
		rel = tlen
	}
	prefix := utf16Prefix(sentence.Text, rel)
	return NewLineColumnPosition(
		sentence.StartLine+CountLineBreaks(prefix),
		ProcessColumnChange(sentence.StartColumn, prefix),
	)
}

// utf16LastIndexOf returns the UTF-16 code-unit index of the last occurrence of r, or -1.
func utf16LastIndexOf(s string, r rune) int {
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

// utf16Prefix returns the prefix of s covering the first n UTF-16 code units
// (Java: s.substring(0, n) with n clamped by caller).
func utf16Prefix(s string, n int) string {
	if n <= 0 {
		return ""
	}
	u := 0
	for i, ch := range s {
		w := 1
		if ch >= 0x10000 {
			w = 2
		}
		if u+w > n {
			return s[:i]
		}
		u += w
		if u == n {
			// include this rune
			// i is byte start; advance past this rune
			return s[:i+len(string(ch))]
		}
	}
	return s
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
