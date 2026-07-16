package de

import (
	"regexp"
	"strings"
)

// GermanNumberInWordFilter ports AbstractNumberInWordFilter suggestion logic
// without a speller: always proposes digit-stripped / 0→o forms when different.
type GermanNumberInWordFilter struct{}

func NewGermanNumberInWordFilter() *GermanNumberInWordFilter {
	return &GermanNumberInWordFilter{}
}

var digitRE = regexp.MustCompile(`[0-9]`)

// Suggestions returns candidate fixes for words containing digits.
func (f *GermanNumberInWordFilter) Suggestions(word string) []string {
	if !digitRE.MatchString(word) {
		return nil
	}
	var out []string
	repl0 := strings.ReplaceAll(word, "0", "o")
	if repl0 != word {
		out = append(out, repl0)
	}
	without := digitRE.ReplaceAllString(word, "")
	if without != "" && without != word {
		// avoid duplicates
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
