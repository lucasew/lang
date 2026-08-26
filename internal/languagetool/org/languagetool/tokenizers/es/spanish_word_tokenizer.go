package es

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// SpanishWordTokenizer ports org.languagetool.tokenizers.es.SpanishWordTokenizer.
type SpanishWordTokenizer struct{}

func NewSpanishWordTokenizer() *SpanishWordTokenizer { return &SpanishWordTokenizer{} }

const wordCharacters = `§©@€£\$_\p{L}\d·\-\x{0300}-\x{036F}\x{00A8}\x{2070}-\x{209F}°%‰‱&\x{FFFD}\x{00AD}\x{00AC}`

var (
	tokenizerPattern = regexp.MustCompile(`[` + wordCharacters + `]+|[^` + wordCharacters + `]`)
	// DECIMAL_* have no UNICODE_CHARACTER_CLASS → ASCII \d only (Java).
	decimalPoint = regexp.MustCompile(`(?i)([\d])\.([\d])`)
	decimalComma = regexp.MustCompile(`(?i)([\d]),([\d])`)
	// Longer ordinal suffixes first. Java ORDINAL_POINT uses \b…\b and \d with
	// UNICODE_CHARACTER_CLASS; Go RE2 \b is ASCII-only and \d is ASCII digits.
	// Digits: \p{Nd} = Java UCC \d. Boundaries: javaUCCWordChar in replaceOrdinalPoint.
	ordinalPoint = regexp.MustCompile(`(?i)(\p{Nd}+)\.(º|ª|er|os|as|o|a)`)
	softHyphen   = regexp.MustCompile(`\x{00AD}`)
	// Java SpanishWordTokenizer.wordsToAdd camel-case hyphen exceptions only.
	javaHyphenExceptions = map[string]bool{
		"mers-cov": true, "mcgraw-hill": true, "sars-cov-2": true, "sars-cov": true,
		"ph-metre": true, "ph-metres": true,
	}
)

// IsTaggedES optional SpanishTagger.INSTANCE.tag(...).isTagged() hook.
// Java keeps hyphen compounds only when SpanishTagger tags them.
// Without a tagger, miss (split hyphens) — do not invent a soft compound lexicon.
var IsTaggedES func(s string) bool

func (w *SpanishWordTokenizer) Tokenize(text string) []string {
	auxText := strings.ReplaceAll(text, "\u2010", "\u002d")
	auxText = strings.ReplaceAll(auxText, "\u2011", "\u002d")
	auxText = decimalPoint.ReplaceAllString(auxText, "${1}xxES_DECIMAL_POINTxx${2}")
	auxText = decimalComma.ReplaceAllString(auxText, "${1}xxES_DECIMAL_COMMAxx${2}")
	auxText = replaceOrdinalPoint(auxText)

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

// javaUCCWordChar ports Java Pattern.UNICODE_CHARACTER_CLASS \w:
// [\p{Alpha}\p{gc=Mn}\p{gc=Me}\p{gc=Mc}\p{Digit}\p{gc=Pc}\p{Join_Control}]
// Used for ORDINAL_POINT \b edges (Java isAlphabetic ≈ letter|Nl; Digit = Nd).
func javaUCCWordChar(r rune) bool {
	if unicode.IsLetter(r) || unicode.Is(unicode.Nl, r) {
		return true
	}
	if unicode.IsDigit(r) { // Nd — Java UCC \d / Digit
		return true
	}
	if unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Me, r) || unicode.Is(unicode.Mc, r) {
		return true
	}
	if unicode.Is(unicode.Pc, r) { // Connector_Punctuation, includes '_'
		return true
	}
	// Join_Control (Cf): U+200C ZWNJ, U+200D ZWJ
	if r == '\u200C' || r == '\u200D' {
		return true
	}
	return false
}

// replaceOrdinalPoint ports Java ORDINAL_POINT.replaceAll("$1xxES_ORDINAL_POINTxx$2")
// with UNICODE_CHARACTER_CLASS \b on both edges: left must be start or non-\w,
// right must be end or non-\w (javaUCCWordChar = Java UCC \w).
func replaceOrdinalPoint(text string) string {
	var b strings.Builder
	last := 0
	for _, loc := range ordinalPoint.FindAllStringSubmatchIndex(text, -1) {
		full0, full1 := loc[0], loc[1]
		// Java leading \b: start or previous rune is non-word (UCC \w)
		if full0 > 0 {
			r, _ := utf8.DecodeLastRuneInString(text[:full0])
			if javaUCCWordChar(r) {
				continue
			}
		}
		// Java trailing \b: end or next rune is non-word (UCC \w)
		if full1 < len(text) {
			r, _ := utf8.DecodeRuneInString(text[full1:])
			if javaUCCWordChar(r) {
				continue
			}
		}
		b.WriteString(text[last:full0])
		g1 := text[loc[2]:loc[3]]
		g2 := text[loc[4]:loc[5]]
		b.WriteString(g1 + "xxES_ORDINAL_POINTxx" + g2)
		last = full1
	}
	b.WriteString(text[last:])
	return b.String()
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
	// Java: SpanishTagger.INSTANCE.tag(...).isTagged() OR equalsIgnoreCase exceptions.
	if isTaggedES(normalized) || javaHyphenExceptions[strings.ToLower(s)] {
		l = append(l, s)
		return l
	}
	// if not found, the word is split on hyphens (keep separators)
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
	return l
}

func isTaggedES(s string) bool {
	// Java: SpanishTagger.INSTANCE.tag(...).isTagged(). Without a tagger, miss
	// (split hyphens) — do not invent a soft compound lexicon.
	if IsTaggedES != nil {
		return IsTaggedES(s)
	}
	return false
}
