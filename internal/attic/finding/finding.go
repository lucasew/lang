package finding

import "strconv"

// SARIF 2.1 result levels (https://docs.oasis-open.org/sarif/sarif/v2.1.0/sarif-v2.1.0.html).
const (
	SeverityError   = "error"
	SeverityWarning = "warning"
	SeverityNote    = "note"
	SeverityNone    = "none"
)

// Finding is one lint issue.
type Finding struct {
	File      string `json:"file"`
	Line      int    `json:"line"`
	Column    int    `json:"column"`
	EndLine   int    `json:"endLine,omitempty"`
	EndColumn int    `json:"endColumn,omitempty"`
	Offset    int    `json:"offset"`    // 0-based rune start in file text
	EndOffset int    `json:"endOffset"` // exclusive
	Rule      string `json:"rule"`
	// Severity is a SARIF level: error | warning | note | none.
	Severity string `json:"severity"`
	// Type is the LanguageTool ITS issue type (e.g. misspelling, whitespace, grammar).
	Type string `json:"type"`
	Message     string   `json:"message"`
	Suggestions []string `json:"suggestions,omitempty"`
	Language    string   `json:"language,omitempty"`
}

// PrimarySuggestion returns the first suggestion, if any.
func (f Finding) PrimarySuggestion() string {
	if len(f.Suggestions) == 0 {
		return ""
	}
	return f.Suggestions[0]
}

// Location formats file:line:col.
func (f Finding) Location() string {
	return f.File + ":" + strconv.Itoa(f.Line) + ":" + strconv.Itoa(f.Column)
}

// SARIFLevel maps a LanguageTool ITS issue type to a SARIF 2.1 reporting level.
//
//   - misspelling / grammar → error (fail CI by default)
//   - style / register / locale-* → note
//   - everything else (whitespace, typographical, duplication, other, …) → warning
func SARIFLevel(issueType string) string {
	switch normalizeType(issueType) {
	case "misspelling", "grammar":
		return SeverityError
	case "style", "register", "locale-violation", "locale-specific-content":
		return SeverityNote
	case "none":
		return SeverityNone
	case SeverityError, SeverityWarning, SeverityNote:
		return normalizeType(issueType)
	default:
		return SeverityWarning
	}
}

// WithType returns normalized ITS type and derived SARIF severity.
func WithType(issueType string) (typ, severity string) {
	typ = normalizeType(issueType)
	if typ == "" {
		typ = "other"
	}
	return typ, SARIFLevel(typ)
}

func normalizeType(s string) string {
	b := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c = c - 'A' + 'a'
		}
		b[i] = c
	}
	return string(b)
}
