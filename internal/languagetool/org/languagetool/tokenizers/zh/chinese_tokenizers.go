package zh

import (
	"strings"
	"unicode"
	"unicode/utf8"
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
		if unicode.IsSpace(r) {
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
type ChineseSentenceTokenizer struct{}

func NewChineseSentenceTokenizer() *ChineseSentenceTokenizer { return &ChineseSentenceTokenizer{} }

func (t *ChineseSentenceTokenizer) Tokenize(text string) []string {
	if text == "" {
		return nil
	}
	// split on Chinese and Latin sentence punctuation
	seps := map[rune]bool{'。': true, '！': true, '？': true, '；': true, '.': true, '!': true, '?': true, '\n': true}
	var out []string
	var buf strings.Builder
	for _, r := range text {
		buf.WriteRune(r)
		if seps[r] {
			s := strings.TrimSpace(buf.String())
			if s != "" {
				out = append(out, s)
			}
			buf.Reset()
		}
	}
	if s := strings.TrimSpace(buf.String()); s != "" {
		out = append(out, s)
	}
	if len(out) == 0 && utf8.RuneCountInString(text) > 0 {
		return []string{text}
	}
	return out
}
