package wikipedia

// PlainTextMapping ports a minimal org.languagetool.dev.wikipedia.PlainTextMapping.
// When original markup equals plain text, positions map 1:1 by line/column.
type PlainTextMapping struct {
	plainText string
	original  string
}

func NewPlainTextMapping(plainText string) *PlainTextMapping {
	return &PlainTextMapping{plainText: plainText, original: plainText}
}

func NewPlainTextMappingWithOriginal(plainText, original string) *PlainTextMapping {
	return &PlainTextMapping{plainText: plainText, original: original}
}

func (m *PlainTextMapping) GetPlainText() string { return m.plainText }
func (m *PlainTextMapping) GetOriginal() string  { return m.original }

// OriginalTextPositionFor returns 1-based line/column for a 1-based plain-text position
// (Java Location is 1-based character index into plain text).
func (m *PlainTextMapping) OriginalTextPositionFor(oneBasedPos int) (line, col int, err error) {
	// walk plain text; oneBasedPos is 1-based index into characters (Java substring style, runes for BMP)
	if oneBasedPos < 1 {
		return 0, 0, fmtError("position must be >= 1")
	}
	line, col = 1, 1
	pos := 0 // 0-based count of characters visited
	for _, ch := range m.plainText {
		if pos+1 == oneBasedPos {
			return line, col, nil
		}
		if ch == '\n' {
			line++
			col = 1
		} else {
			col++
		}
		pos++
	}
	// position at end (past last char) — Java uses toPos+1 for exclusive end
	if pos+1 == oneBasedPos {
		return line, col, nil
	}
	return 0, 0, fmtError("position out of range")
}

type mappingError string

func (e mappingError) Error() string { return string(e) }
func fmtError(s string) error        { return mappingError(s) }
