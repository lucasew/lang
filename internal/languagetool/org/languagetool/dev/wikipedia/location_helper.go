package wikipedia

import "fmt"

// AbsolutePositionFor maps 1-based line/column to a 0-based character offset (Java LocationHelper simplified).
// Wiki markup ignore levels are deferred; plain text line/col works for the LocationHelperTest fixtures.
func AbsolutePositionFor(line, column int, text string) (int, error) {
	curLine, curCol := 1, 1
	pos := 0
	for _, ch := range text {
		if curLine == line && curCol == column {
			return pos, nil
		}
		if ch == '\n' {
			curLine++
			curCol = 1
		} else {
			curCol++
		}
		pos++
	}
	if curLine == line && curCol == column {
		return pos, nil
	}
	return -1, fmt.Errorf("could not find location %d:%d in text (max %d:%d)", line, column, curLine, curCol)
}
