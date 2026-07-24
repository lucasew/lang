package zh

import (
	"strings"
	"unicode"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// ChineseWordTokenizer ports tokenizers.zh.ChineseWordTokenizer.
//
// Java uses HanLP and returns each Term as Term.toString() → "surface/pos".
// Until HanLP (or an equivalent) is wired, this does not invent a soft lexicon
// from grammar packs: it falls back to per-rune surfaces with POS "x".
// See inspiration/languagetool language-modules/zh ChineseWordTokenizer.
type ChineseWordTokenizer struct {
	// Segment optional custom segmenter (surfaces only).
	Segment func(text string) []string
}

func NewChineseWordTokenizer() *ChineseWordTokenizer { return &ChineseWordTokenizer{} }

func (t *ChineseWordTokenizer) Tokenize(text string) []string {
	if t != nil && t.Segment != nil {
		return encodeChineseTerms(t.Segment(text))
	}
	// No HanLP: do not invent multi-character soft lexicons.
	return encodeChineseTerms(segmentRunes(text))
}

// segmentRunes is the incomplete no-HanLP path: CJK (and other non-Latin) chars
// stay per-rune; consecutive Latin/digit runs stay whole (HanLP does the same
// for "world"). Does not invent multi-character Chinese words.
func segmentRunes(text string) []string {
	if text == "" {
		return nil
	}
	var out []string
	var latin strings.Builder
	flushLatin := func() {
		if latin.Len() > 0 {
			out = append(out, latin.String())
			latin.Reset()
		}
	}
	for _, r := range text {
		// Incomplete no-HanLP path: drop Character.isWhitespace runs (not Go unicode.IsSpace —
		// NBSP is not whitespace in Java Character.isWhitespace and stays as surface).
		if tools.CharacterIsWhitespace(r) {
			flushLatin()
			continue
		}
		// Keep Latin letters and digits as multi-char runs (ASCII/Latin-1 style).
		if isLatinOrDigit(r) {
			latin.WriteRune(r)
			continue
		}
		flushLatin()
		out = append(out, string(r))
	}
	flushLatin()
	return out
}

func isLatinOrDigit(r rune) bool {
	if unicode.IsDigit(r) {
		return true
	}
	// Basic Latin + Latin-1 supplement letters (HanLP keeps "world" as one term).
	if r <= 0x024F && unicode.IsLetter(r) {
		return true
	}
	// Common Latin extensions used in mixed ZH text
	if r >= 0x1E00 && r <= 0x1EFF && unicode.IsLetter(r) {
		return true
	}
	return false
}

// encodeChineseTerms maps surfaces to Java HanLP Term.toString form "surface/pos".
// Without HanLP, POS is "x" (unknown) — not invented soft tags from grammar packs.
func encodeChineseTerms(surfaces []string) []string {
	if len(surfaces) == 0 {
		return nil
	}
	out := make([]string, 0, len(surfaces))
	for _, s := range surfaces {
		if s == "" {
			continue
		}
		out = append(out, s+"/x")
	}
	return out
}

// ChineseSentenceTokenizer ports tokenizers.zh.ChineseSentenceTokenizer.
//
// Java walks text by char (UTF-16 unit), emits Character.isWhitespace runs as
// their own tokens, and runs HanLP SentencesUtil.toSentenceList on each
// non-whitespace chunk so match positions stay aligned.
//
// Twin: inspiration/languagetool/.../tokenizers/zh/ChineseSentenceTokenizer.java
type ChineseSentenceTokenizer struct{}

func NewChineseSentenceTokenizer() *ChineseSentenceTokenizer { return &ChineseSentenceTokenizer{} }

func (t *ChineseSentenceTokenizer) Tokenize(text string) []string {
	// Java:
	//   for (int i = 0; i < text.length(); i++) {
	//     if (Character.isWhitespace(text.charAt(i))) { ... }
	//   }
	// Buffers hold UTF-16 units so surrogate pairs match Java StringBuilder.
	var result []string
	var whitespace []uint16
	var nonWhitespace []uint16
	for _, u := range utf16.Encode([]rune(text)) {
		ch := rune(u)
		// Java Character.isWhitespace — not Go unicode.IsSpace (NBSP differs).
		if tools.CharacterIsWhitespace(ch) {
			if len(nonWhitespace) > 0 {
				result = append(result, sentencesUtilToSentenceList(utf16UnitsToString(nonWhitespace))...)
				nonWhitespace = nonWhitespace[:0]
			}
			whitespace = append(whitespace, u)
			continue
		}
		if len(whitespace) > 0 {
			result = append(result, utf16UnitsToString(whitespace))
			whitespace = whitespace[:0]
		}
		nonWhitespace = append(nonWhitespace, u)
	}
	if len(whitespace) > 0 {
		result = append(result, utf16UnitsToString(whitespace))
	}
	if len(nonWhitespace) > 0 {
		result = append(result, sentencesUtilToSentenceList(utf16UnitsToString(nonWhitespace))...)
	}
	return result
}

// SetSingleLineBreaksMarksParagraph ports ChineseSentenceTokenizer note:
// no effect for Chinese (Java empty override).
func (t *ChineseSentenceTokenizer) SetSingleLineBreaksMarksParagraph(lineBreakParagraphs bool) {
}

// SingleLineBreaksMarksPara ports ChineseSentenceTokenizer.singleLineBreaksMarksPara → false.
func (t *ChineseSentenceTokenizer) SingleLineBreaksMarksPara() bool {
	return false
}
