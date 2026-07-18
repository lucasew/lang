package es

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// SpanishWordTokenizer ports org.languagetool.tokenizers.es.SpanishWordTokenizer.
type SpanishWordTokenizer struct{}

func NewSpanishWordTokenizer() *SpanishWordTokenizer { return &SpanishWordTokenizer{} }

const wordCharacters = `§©@€£\$_\p{L}\d·\-\x{0300}-\x{036F}\x{00A8}\x{2070}-\x{209F}°%‰‱&\x{FFFD}\x{00AD}\x{00AC}`

var (
	tokenizerPattern = regexp.MustCompile(`[` + wordCharacters + `]+|[^` + wordCharacters + `]`)
	decimalPoint     = regexp.MustCompile(`(?i)([\d])\.([\d])`)
	decimalComma     = regexp.MustCompile(`(?i)([\d]),([\d])`)
	// Longer ordinal suffixes first. No trailing \b: Go's \b is ASCII-only and
	// fails after º/ª (Java uses UNICODE_CHARACTER_CLASS).
	ordinalPoint = regexp.MustCompile(`(?i)\b([\d]+)\.(º|ª|er|os|as|o|a)`)
	softHyphen       = regexp.MustCompile(`\x{00AD}`)
	// Soft: keep only known dictionary-like compounds (Java keeps when tagged).
	// Default is split so grammar patterns with explicit "-" tokens can match.
	hyphenExceptions = map[string]bool{
		"mers-cov": true, "mcgraw-hill": true, "sars-cov-2": true, "sars-cov": true,
		"ph-metre": true, "ph-metres": true,
		"e-mails": true, "e-mail": true, "best-seller": true, "best-sellers": true,
		"covid-19": true, "al-ándalus": true, "al-andalus": true,
	}
)

func (w *SpanishWordTokenizer) Tokenize(text string) []string {
	auxText := strings.ReplaceAll(text, "\u2010", "\u002d")
	auxText = strings.ReplaceAll(auxText, "\u2011", "\u002d")
	auxText = decimalPoint.ReplaceAllString(auxText, "${1}xxES_DECIMAL_POINTxx${2}")
	auxText = decimalComma.ReplaceAllString(auxText, "${1}xxES_DECIMAL_COMMAxx${2}")
	auxText = ordinalPoint.ReplaceAllString(auxText, "${1}xxES_ORDINAL_POINTxx${2}")

	var l []string
	for _, loc := range tokenizerPattern.FindAllStringIndex(auxText, -1) {
		s := auxText[loc[0]:loc[1]]
		if len(l) > 0 {
			r, size := utf8.DecodeRuneInString(s)
			if size == len(s) && r >= 0xFE00 && r <= 0xFE0F {
				l[len(l)-1] = l[len(l)-1] + s
				continue
			}
		}
		s = strings.ReplaceAll(s, "xxES_DECIMAL_POINTxx", ".")
		s = strings.ReplaceAll(s, "xxES_DECIMAL_COMMAxx", ",")
		s = strings.ReplaceAll(s, "xxES_ORDINAL_POINTxx", ".")
		l = append(l, wordsToAddES(s)...)
	}
	return tokenizers.JoinEMailsAndUrls(l)
}

func wordsToAddES(s string) []string {
	var l []string
	if s == "" {
		return l
	}
	if !strings.Contains(s, "-") {
		l = append(l, s)
		return l
	}
	normalized := softHyphen.ReplaceAllString(s, "")
	normalized = strings.ReplaceAll(normalized, "’", "'")
	if isTaggedES(normalized) || hyphenExceptions[strings.ToLower(s)] {
		l = append(l, s)
	} else {
		// split on hyphen, keep delims
		var cur strings.Builder
		for _, r := range s {
			if r == '-' {
				if cur.Len() > 0 {
					l = append(l, cur.String())
					cur.Reset()
				}
				l = append(l, "-")
			} else {
				cur.WriteRune(r)
			}
		}
		if cur.Len() > 0 {
			l = append(l, cur.String())
		}
	}
	return l
}

func isTaggedES(s string) bool {
	// Soft path without SpanishTagger: do not keep arbitrary hyphen compounds.
	// Java only keeps hyphenated forms when the dictionary tags them; untagged
	// forms are split so rules like PREFIJO_CUASI / DEJA_VU can match.
	_ = s
	return false
}
