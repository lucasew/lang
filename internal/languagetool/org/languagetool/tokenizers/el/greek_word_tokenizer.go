package el

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// GreekWordTokenizer ports tokenizers.el.GreekWordTokenizer (simplified).
type GreekWordTokenizer struct{}

func NewGreekWordTokenizer() *GreekWordTokenizer { return &GreekWordTokenizer{} }

func (t *GreekWordTokenizer) Tokenize(text string) []string {
	if text == "" {
		return nil
	}
	var out []string
	var buf strings.Builder
	flush := func() {
		if buf.Len() > 0 {
			out = append(out, buf.String())
			buf.Reset()
		}
	}
	for _, r := range text {
		if unicode.IsLetter(r) || r == '\'' || r == '΄' || r == '’' {
			buf.WriteRune(r)
			continue
		}
		flush()
		if !unicode.IsSpace(r) || r == '\n' {
			out = append(out, string(r))
		} else if r == ' ' || r == '\t' {
			out = append(out, string(r))
		}
	}
	flush()
	if len(out) == 0 && utf8.RuneCountInString(text) > 0 {
		return []string{text}
	}
	return out
}
