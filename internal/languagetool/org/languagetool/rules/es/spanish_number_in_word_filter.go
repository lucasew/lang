package es

import (
	"regexp"
	"strings"
)

// SpanishNumberInWordFilter ports AbstractNumberInWordFilter suggestion logic
// without a speller: always proposes digit-stripped / 0→o forms when different.
type SpanishNumberInWordFilter struct{}

func NewSpanishNumberInWordFilter() *SpanishNumberInWordFilter {
	return &SpanishNumberInWordFilter{}
}

var esDigitRE = regexp.MustCompile(`[0-9]`)

// Suggestions returns candidate fixes for words containing digits.
func (f *SpanishNumberInWordFilter) Suggestions(word string) []string {
	if !esDigitRE.MatchString(word) {
		return nil
	}
	var out []string
	repl0 := strings.ReplaceAll(word, "0", "o")
	if repl0 != word {
		out = append(out, repl0)
	}
	without := esDigitRE.ReplaceAllString(word, "")
	if without != "" && without != word {
		dup := false
		for _, s := range out {
			if s == without {
				dup = true
				break
			}
		}
		if !dup {
			out = append(out, without)
		}
	}
	return out
}
