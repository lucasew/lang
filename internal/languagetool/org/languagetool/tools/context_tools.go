package tools

import (
	"strings"
	"unicode/utf16"
)

// ContextTools ports org.languagetool.tools.ContextTools — mark errors in text context.
type ContextTools struct {
	contextSize      int
	escapeHTML       bool
	errorMarkerStart string
	errorMarkerEnd   string
}

func NewContextTools() *ContextTools {
	return &ContextTools{
		contextSize:      40,
		escapeHTML:       true,
		errorMarkerStart: `<b><font bgcolor="#ff8b8b">`,
		errorMarkerEnd:   `</font></b>`,
	}
}

func (c *ContextTools) SetContextSize(n int)          { c.contextSize = n }
func (c *ContextTools) SetEscapeHtml(v bool)           { c.escapeHTML = v }
func (c *ContextTools) SetErrorMarker(start, end string) {
	c.errorMarkerStart = start
	c.errorMarkerEnd = end
}
func (c *ContextTools) SetErrorMarkerStart(s string) { c.errorMarkerStart = s }
func (c *ContextTools) SetErrorMarkerEnd(s string)   { c.errorMarkerEnd = s }

// GetContext ports getContext — HTML (optional escape) context with error markers.
// Positions are UTF-16 code unit indices (Java String offsets).
func (c *ContextTools) GetContext(fromPos, toPos int, contents string) string {
	startContent := fromPos - c.contextSize
	prefix := "..."
	postfix := "..."
	if startContent < 0 {
		prefix = ""
		startContent = 0
	}
	endContent := toPos + c.contextSize
	textLength := utf16LenTools(contents)
	if endContent > textLength {
		postfix = ""
		endContent = textLength
	}
	// build context string plus marker
	chunk := utf16Substr(contents, startContent, endContent)
	chunk = strings.ReplaceAll(chunk, "\n", " ")
	markerStr := getMarker(fromPos, toPos, startContent, endContent, prefix)
	result := prefix + chunk + postfix
	startMark := strings.IndexByte(markerStr, '^')
	endMark := strings.LastIndexByte(markerStr, '^')
	if startMark < 0 {
		return result
	}
	if c.escapeHTML {
		escapedErrorPart := strings.ReplaceAll(EscapeHTML(result[startMark:endMark+1]), " ", "&nbsp;")
		result = EscapeHTML(result[:startMark]) +
			c.errorMarkerStart +
			escapedErrorPart +
			c.errorMarkerEnd + EscapeHTML(result[endMark+1:])
	} else {
		result = result[:startMark] + c.errorMarkerStart +
			result[startMark:endMark+1] + c.errorMarkerEnd +
			result[endMark+1:]
	}
	return result
}

// GetPlainTextContext ports getPlainTextContext — uses ^ markers on a second line.
func (c *ContextTools) GetPlainTextContext(fromPos, toPos int, contents string) string {
	startContent := fromPos - c.contextSize
	prefix := "..."
	postfix := "..."
	if startContent < 0 {
		prefix = ""
		startContent = 0
	}
	endContent := toPos + c.contextSize
	textLength := utf16LenTools(contents)
	if endContent > textLength {
		postfix = ""
		endContent = textLength
	}
	chunk := utf16Substr(contents, startContent, endContent)
	chunk = strings.ReplaceAll(chunk, "\n", " ")
	chunk = strings.ReplaceAll(chunk, "\r", " ")
	chunk = strings.ReplaceAll(chunk, "\t", " ")
	return prefix + chunk + postfix + "\n" +
		getMarker(fromPos, toPos, startContent, endContent, prefix)
}

func getMarker(fromPos, toPos, startContent, endContent int, prefix string) string {
	spacesBefore := prefixLenUTF16(prefix) + fromPos - startContent
	carets := toPos - fromPos
	spacesAfter := endContent - toPos
	if spacesBefore < 0 {
		spacesBefore = 0
	}
	if carets < 0 {
		carets = 0
	}
	if spacesAfter < 0 {
		spacesAfter = 0
	}
	return strings.Repeat(" ", spacesBefore) + strings.Repeat("^", carets) + strings.Repeat(" ", spacesAfter)
}

func prefixLenUTF16(prefix string) int { return utf16LenTools(prefix) }

func utf16LenTools(s string) int {
	return len(utf16.Encode([]rune(s)))
}

func utf16Substr(s string, from, to int) string {
	u := utf16.Encode([]rune(s))
	if from < 0 {
		from = 0
	}
	if to > len(u) {
		to = len(u)
	}
	if from >= to {
		return ""
	}
	return string(utf16.Decode(u[from:to]))
}
