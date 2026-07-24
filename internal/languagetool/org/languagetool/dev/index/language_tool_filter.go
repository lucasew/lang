package index

import (
	"strings"
	"unicode"
)

// LanguageToolFilter ports dev.index.LanguageToolFilter surface —
// tokenizes text for indexing without full LT analysis (WordTokenizer soft).
type LanguageToolFilter struct {
	// LowerCase folds tokens.
	LowerCase bool
}

func NewLanguageToolFilter(lower bool) *LanguageToolFilter {
	return &LanguageToolFilter{LowerCase: lower}
}

// Tokenize splits on non-letters/digits (soft index tokens).
func (f *LanguageToolFilter) Tokenize(text string) []string {
	var out []string
	var b strings.Builder
	flush := func() {
		if b.Len() == 0 {
			return
		}
		tok := b.String()
		b.Reset()
		if f != nil && f.LowerCase {
			tok = strings.ToLower(tok)
		}
		out = append(out, tok)
	}
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else {
			flush()
		}
	}
	flush()
	return out
}
