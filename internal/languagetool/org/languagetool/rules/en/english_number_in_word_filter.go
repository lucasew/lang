package en

import (
	"regexp"
	"strings"
)

// EnglishNumberInWordFilter ports AbstractNumberInWordFilter suggestion logic.
type EnglishNumberInWordFilter struct{}

func NewEnglishNumberInWordFilter() *EnglishNumberInWordFilter {
	return &EnglishNumberInWordFilter{}
}

var enDigitRE = regexp.MustCompile(`[0-9]`)

func (f *EnglishNumberInWordFilter) Suggestions(word string) []string {
	if !enDigitRE.MatchString(word) {
		return nil
	}
	var out []string
	repl0 := strings.ReplaceAll(word, "0", "o")
	if repl0 != word {
		out = append(out, repl0)
	}
	without := enDigitRE.ReplaceAllString(word, "")
	if without != "" && without != word {
		found := false
		for _, s := range out {
			if s == without {
				found = true
				break
			}
		}
		if !found {
			out = append(out, without)
		}
	}
	return out
}
