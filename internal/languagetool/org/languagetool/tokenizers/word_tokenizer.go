package tokenizers

import (
	"unicode/utf16"
)

// TOKENIZING_CHARACTERS — subset from WordTokenizer.java (whitespace + common punct).
const tokenizing = " \t\n\r\u00A0\u200B\uFEFF\u2060" +
	".,;:?!…'\"„“”»«()[]{}<>/\\|*+=~`@#%^&" +
	"\u2000\u2001\u2002\u2003\u2004\u2005\u2006\u2007\u2008\u2009\u200A"

// WordTokenizer ports org.languagetool.tokenizers.WordTokenizer (simplified but keeps
// whitespace as separate single-char tokens when they are tokenizing chars).
type WordTokenizer struct{}

func NewWordTokenizer() *WordTokenizer { return &WordTokenizer{} }

// Tokenize returns tokens; positions are UTF-16 code unit indices (Java String offsets).
func (w *WordTokenizer) Tokenize(text string) []string {
	if text == "" {
		return nil
	}
	var out []string
	var cur []rune
	flush := func() {
		if len(cur) > 0 {
			out = append(out, string(cur))
			cur = nil
		}
	}
	for _, r := range text {
		if isTokenizing(r) {
			flush()
			out = append(out, string(r))
		} else {
			cur = append(cur, r)
		}
	}
	flush()
	return out
}

func isTokenizing(r rune) bool {
	for _, t := range tokenizing {
		if r == t {
			return true
		}
	}
	return false
}

// UTF16Len returns Java String.length() equivalent.
func UTF16Len(s string) int {
	return len(utf16.Encode([]rune(s)))
}

// BuildPositions returns start offset (UTF-16) for each token when concatenated in order.
func BuildPositions(tokens []string) []int {
	pos := make([]int, len(tokens))
	p := 0
	for i, t := range tokens {
		pos[i] = p
		p += UTF16Len(t)
	}
	return pos
}
