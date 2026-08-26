package tokenizers

import "strings"

// SimpleCompoundWordTokenizer splits on hyphens/dashes (minimal CompoundWordTokenizer).
type SimpleCompoundWordTokenizer struct{}

func NewSimpleCompoundWordTokenizer() *SimpleCompoundWordTokenizer {
	return &SimpleCompoundWordTokenizer{}
}

func (t *SimpleCompoundWordTokenizer) Tokenize(text string) []string {
	if text == "" {
		return nil
	}
	// split on common dash characters but keep non-empty parts
	parts := strings.FieldsFunc(text, func(r rune) bool {
		return r == '-' || r == '‐' || r == '‑' || r == '–' || r == '—'
	})
	var out []string
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		return []string{text}
	}
	return out
}

var _ CompoundWordTokenizer = (*SimpleCompoundWordTokenizer)(nil)
