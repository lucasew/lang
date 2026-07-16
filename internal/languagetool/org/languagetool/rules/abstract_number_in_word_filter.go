package rules

import (
	"regexp"
	"strings"
)

// NumberInWordFilter ports AbstractNumberInWordFilter surface suggestions
// (0→o and digit-stripped forms). Speller gating is left to callers.
type NumberInWordFilter struct{}

func NewNumberInWordFilter() *NumberInWordFilter {
	return &NumberInWordFilter{}
}

var numberInWordDigitRE = regexp.MustCompile(`[0-9]`)

// Suggestions returns candidate fixes for words containing digits.
func (f *NumberInWordFilter) Suggestions(word string) []string {
	if !numberInWordDigitRE.MatchString(word) {
		return nil
	}
	var out []string
	repl0 := strings.ReplaceAll(word, "0", "o")
	if repl0 != word {
		out = append(out, repl0)
	}
	without := numberInWordDigitRE.ReplaceAllString(word, "")
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
