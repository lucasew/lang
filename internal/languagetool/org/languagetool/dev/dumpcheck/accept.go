package dumpcheck

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

const (
	MinSentenceLength     = 10
	MinSentenceTokenCount = 4
	MaxSentenceLength     = 300
)

// AcceptSentence ports SentenceSource.acceptSentence length/token filters.
// Optional acceptPattern (regexp) mirrors Java Pattern filter.
func AcceptSentence(sentence string, acceptPattern *regexp.Regexp) bool {
	if acceptPattern != nil && !acceptPattern.MatchString(sentence) {
		return false
	}
	trim := strings.TrimSpace(sentence)
	if len(trim) < MinSentenceLength || len(trim) > MaxSentenceLength {
		return false
	}
	return countTokens(trim) >= MinSentenceTokenCount
}

func countTokens(sentence string) int {
	wt := tokenizers.NewWordTokenizer()
	n := 0
	for _, t := range wt.Tokenize(sentence) {
		if strings.TrimSpace(t) != "" {
			n++
		}
	}
	// fallback if tokenizer empty: whitespace words
	if n == 0 {
		for _, f := range strings.FieldsFunc(sentence, unicode.IsSpace) {
			if f != "" {
				n++
			}
		}
	}
	return n
}
