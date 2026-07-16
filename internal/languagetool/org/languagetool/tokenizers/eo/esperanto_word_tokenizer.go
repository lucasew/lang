package eo

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// EsperantoWordTokenizer ports org.languagetool.tokenizers.eo.EsperantoWordTokenizer.
type EsperantoWordTokenizer struct {
	base *tokenizers.WordTokenizer
}

func NewEsperantoWordTokenizer() *EsperantoWordTokenizer {
	return &EsperantoWordTokenizer{base: tokenizers.NewWordTokenizer()}
}

const (
	apos1 = "\u0001\u0001EO@APOS1\u0001\u0001"
	apos2 = "\u0001\u0001EO@APOS2\u0001\u0001"
)

func (w *EsperantoWordTokenizer) Tokenize(text string) []string {
	replaced := protectEsperantoApostrophes(text)
	tokenList := w.base.Tokenize(replaced)
	var tokens []string
	for i := 0; i < len(tokenList); i++ {
		word := tokenList[i]
		if strings.HasSuffix(word, apos2) {
			// Skip the next spurious white space.
			if i+1 < len(tokenList) {
				i++
			}
		}
		word = strings.ReplaceAll(word, apos1, "'")
		word = strings.ReplaceAll(word, apos2, "'")
		tokens = append(tokens, word)
	}
	return tokens
}

func protectEsperantoApostrophes(text string) string {
	// Emulate Java lookaround patterns without RE2 lookbehind/lookahead.
	var b strings.Builder
	runes := []rune(text)
	isEO := func(r rune) bool {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			return true
		}
		switch r {
		case 'ĉ', 'ĝ', 'ĥ', 'ĵ', 'ŝ', 'ŭ', 'Ĉ', 'Ĝ', 'Ĥ', 'Ĵ', 'Ŝ', 'Ŭ':
			return true
		}
		return false
	}
	isWordStart := func(i int) bool {
		if i == 0 {
			return true
		}
		prev := runes[i-1]
		return !isEO(prev) && prev != '\''
	}
	i := 0
	for i < len(runes) {
		if !isEO(runes[i]) {
			b.WriteRune(runes[i])
			i++
			continue
		}
		if !isWordStart(i) {
			b.WriteRune(runes[i])
			i++
			continue
		}
		j := i
		for j < len(runes) && isEO(runes[j]) {
			j++
		}
		if j < len(runes) && runes[j] == '\'' {
			if j+1 >= len(runes) || (!isEO(runes[j+1]) && runes[j+1] != '-') {
				// pattern1: word' not followed by letter/hyphen
				b.WriteString(string(runes[i:j]))
				b.WriteString(apos1)
				i = j + 1
				continue
			}
			if isEO(runes[j+1]) || runes[j+1] == '-' {
				// pattern2: word' followed by letter/hyphen → insert space after marker
				b.WriteString(string(runes[i:j]))
				b.WriteString(apos2)
				b.WriteByte(' ')
				i = j + 1
				continue
			}
		}
		b.WriteString(string(runes[i:j]))
		i = j
	}
	return b.String()
}
