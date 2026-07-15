package finding

// Finding is one lint issue (1:1 field set with LanguageTool matches).
type Finding struct {
	File        string   `json:"file"`
	Line        int      `json:"line"`
	Column      int      `json:"column"`
	EndLine     int      `json:"endLine,omitempty"`
	EndColumn   int      `json:"endColumn,omitempty"`
	Offset      int      `json:"offset"`    // 0-based byte/rune start in file text
	EndOffset   int      `json:"endOffset"` // exclusive
	Rule        string   `json:"rule"`
	Severity    string   `json:"severity"` // LT ITS issue type (e.g. whitespace, misspelling)
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
