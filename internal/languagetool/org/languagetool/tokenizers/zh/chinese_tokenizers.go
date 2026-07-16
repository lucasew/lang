package zh

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// ChineseWordTokenizer ports tokenizers.zh.ChineseWordTokenizer (character-level fallback).
// Real Chinese segmentation (ICTCLAS/jieba) is deferred.
type ChineseWordTokenizer struct {
	// Segment optional custom segmenter.
	Segment func(text string) []string
}

func NewChineseWordTokenizer() *ChineseWordTokenizer { return &ChineseWordTokenizer{} }

func (t *ChineseWordTokenizer) Tokenize(text string) []string {
	if t != nil && t.Segment != nil {
		return t.Segment(text)
	}
	if text == "" {
		return nil
	}
	// split CJK runs into characters; keep non-CJK runs as words
	var out []string
	var buf strings.Builder
	flush := func() {
		if buf.Len() > 0 {
			out = append(out, buf.String())
			buf.Reset()
		}
	}
	for _, r := range text {
		if unicode.Is(unicode.Han, r) {
			flush()
			out = append(out, string(r))
			continue
		}
		if unicode.IsSpace(r) {
			flush()
			out = append(out, string(r))
			continue
		}
		buf.WriteRune(r)
	}
	flush()
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
