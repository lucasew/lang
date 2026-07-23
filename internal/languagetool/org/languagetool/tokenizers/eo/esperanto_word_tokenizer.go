package eo

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// EsperantoWordTokenizer ports org.languagetool.tokenizers.eo.EsperantoWordTokenizer.
type EsperantoWordTokenizer struct {
	base *tokenizers.WordTokenizer
}

// NewEsperantoWordTokenizer ports the Java default constructor.
func NewEsperantoWordTokenizer() *EsperantoWordTokenizer {
	return &EsperantoWordTokenizer{base: tokenizers.NewWordTokenizer()}
}

// Markers used while protecting Esperanto apostrophes (same as Java).
const (
	// Java: "\u0001\u0001EO@APOS1\u0001\u0001"
	eoApos1 = "\u0001\u0001EO@APOS1\u0001\u0001"
	// Java: "\u0001\u0001EO@APOS2\u0001\u0001"
	eoApos2 = "\u0001\u0001EO@APOS2\u0001\u0001"
)

// eoLetters is the Java character class [a-zA-ZĉĝĥĵŝŭĈĜĤĴŜŬ].
func eoLetter(r rune) bool {
	if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
		return true
	}
	switch r {
	case 'ĉ', 'ĝ', 'ĥ', 'ĵ', 'ŝ', 'ŭ', 'Ĉ', 'Ĝ', 'Ĥ', 'Ĵ', 'Ŝ', 'Ŭ':
		return true
	}
	return false
}

// javaASCIIWordChar is Java's default \w for \b: [a-zA-Z0-9_] (no UNICODE_CHARACTER_CLASS).
func javaASCIIWordChar(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') ||
		(r >= '0' && r <= '9') || r == '_'
}

// javaWordBoundaryAt ports Java default \b at rune index i.
func javaWordBoundaryAt(runes []rune, i int) bool {
	left := false
	if i > 0 {
		left = javaASCIIWordChar(runes[i-1])
	}
	right := false
	if i < len(runes) {
		right = javaASCIIWordChar(runes[i])
	}
	return left != right
}

// Tokenize ports EsperantoWordTokenizer.tokenize.
//
// Java applies:
//
//	PATTERN_1 = (?<!')\b([a-zA-ZĉĝĥĵŝŭĈĜĤĴŜŬ]+)'(?![a-zA-ZĉĝĥĵŝŭĈĜĤĴŜŬ-])
//	PATTERN_2 = (?<!')\b([a-zA-ZĉĝĥĵŝŭĈĜĤĴŜŬ]+)'(?=[a-zA-ZĉĝĥĵŝŭĈĜĤĴŜŬ-])
//
// then super.tokenize, then restores apostrophes and drops the spurious space
// inserted by PATTERN_2. RE2 cannot express lookaround, so the patterns are
// emulated with an equivalent left-to-right scan (bug-for-bug with Java \b).
func (w *EsperantoWordTokenizer) Tokenize(text string) []string {
	// TODO(java): find a cleaner implementation, this is a hack
	replaced := replaceAllEOAposPattern1(text)
	replaced = replaceAllEOAposPattern2(replaced)
	tokenList := w.base.Tokenize(replaced)
	tokens := make([]string, 0, len(tokenList))
	for i := 0; i < len(tokenList); i++ {
		word := tokenList[i]
		if strings.HasSuffix(word, eoApos2) {
			// Java: itr.next() — skip the next spurious white space.
			if i+1 < len(tokenList) {
				i++
			}
		}
		word = strings.ReplaceAll(word, eoApos1, "'")
		word = strings.ReplaceAll(word, eoApos2, "'")
		tokens = append(tokens, word)
	}
	return tokens
}

// replaceAllEOAposPattern1 ports PATTERN_1.replaceAll("$1\u0001\u0001EO@APOS1\u0001\u0001").
func replaceAllEOAposPattern1(text string) string {
	return replaceAllEOApos(text, true)
}

// replaceAllEOAposPattern2 ports PATTERN_2.replaceAll("$1\u0001\u0001EO@APOS2\u0001\u0001 ").
func replaceAllEOAposPattern2(text string) string {
	return replaceAllEOApos(text, false)
}

// replaceAllEOApos emulates one Java Matcher.replaceAll for PATTERN_1 (pattern1=true)
// or PATTERN_2 (pattern1=false).
func replaceAllEOApos(text string, pattern1 bool) string {
	runes := []rune(text)
	var b strings.Builder
	b.Grow(len(text))
	i := 0
	for i < len(runes) {
		// Fast path: not start of [eo letters]+
		if !eoLetter(runes[i]) {
			b.WriteRune(runes[i])
			i++
			continue
		}
		// (?<!')
		if i > 0 && runes[i-1] == '\'' {
			b.WriteRune(runes[i])
			i++
			continue
		}
		// \b (Java default, ASCII \w)
		if !javaWordBoundaryAt(runes, i) {
			b.WriteRune(runes[i])
			i++
			continue
		}
		// ([a-zA-ZĉĝĥĵŝŭĈĜĤĴŜŬ]+)
		j := i
		for j < len(runes) && eoLetter(runes[j]) {
			j++
		}
		// required trailing '
		if j >= len(runes) || runes[j] != '\'' {
			// No match at i; Java advances one position.
			b.WriteRune(runes[i])
			i++
			continue
		}
		// Lookahead: (?![eo-]) for pattern1, (?=[eo-]) for pattern2.
		letterOrHyphen := false
		if j+1 < len(runes) {
			n := runes[j+1]
			letterOrHyphen = eoLetter(n) || n == '-'
		}
		if pattern1 {
			if letterOrHyphen {
				b.WriteRune(runes[i])
				i++
				continue
			}
			// $1 + EO@APOS1
			b.WriteString(string(runes[i:j]))
			b.WriteString(eoApos1)
			i = j + 1
			continue
		}
		// pattern2
		if !letterOrHyphen {
			b.WriteRune(runes[i])
			i++
			continue
		}
		// $1 + EO@APOS2 + ' '  (trailing space is intentional)
		b.WriteString(string(runes[i:j]))
		b.WriteString(eoApos2)
		b.WriteByte(' ')
		i = j + 1
	}
	return b.String()
}
