package br

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// BretonWordTokenizer ports org.languagetool.tokenizers.br.BretonWordTokenizer.
type BretonWordTokenizer struct {
	base *tokenizers.WordTokenizer
}

func NewBretonWordTokenizer() *BretonWordTokenizer {
	return &BretonWordTokenizer{base: tokenizers.NewWordTokenizer()}
}

var (
	repl1 = regexp.MustCompile(`([Cc])['’‘ʼ]([Hh])`)
	repl2 = regexp.MustCompile(`(\p{L})['’‘ʼ]`)
	apos  = "\u0001\u0001BR@APOS\u0001\u0001"
)

func (w *BretonWordTokenizer) Tokenize(text string) []string {
	replaced := repl1.ReplaceAllString(text, "$1"+apos+"$2")
	replaced = repl2.ReplaceAllString(replaced, "$1"+apos+" ")
	tokenList := w.base.Tokenize(replaced)
	var tokens []string
	for i := 0; i < len(tokenList); i++ {
		word := strings.ReplaceAll(tokenList[i], apos, "’")
		tokens = append(tokens, word)
		if word != "’" && strings.HasSuffix(word, "’") {
			// Skip next spurious white space.
			if i+1 < len(tokenList) {
				i++
			}
		}
	}
	return tokens
}
