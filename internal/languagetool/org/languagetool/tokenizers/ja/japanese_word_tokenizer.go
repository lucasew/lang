package ja

import (
	"strings"
	"unicode"
)

// JapaneseWordTokenizer ports tokenizers.ja.JapaneseWordTokenizer.
// Full morphological analysis (Kuromoji) deferred — falls back to script-run splitting.
type JapaneseWordTokenizer struct {
	Segment func(text string) []string
}

func NewJapaneseWordTokenizer() *JapaneseWordTokenizer { return &JapaneseWordTokenizer{} }

func (t *JapaneseWordTokenizer) Tokenize(text string) []string {
	if t != nil && t.Segment != nil {
		return t.Segment(text)
	}
	if text == "" {
		return nil
	}
	var out []string
	var buf strings.Builder
	var lastClass int // 0 other, 1 kana, 2 kanji, 3 latin
	flush := func() {
		if buf.Len() > 0 {
			out = append(out, buf.String())
			buf.Reset()
		}
	}
	classOf := func(r rune) int {
		switch {
		case unicode.In(r, unicode.Hiragana, unicode.Katakana):
			return 1
		case unicode.Is(unicode.Han, r):
			return 2
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			return 3
		default:
			return 0
		}
	}
	for _, r := range text {
		if unicode.IsSpace(r) {
			flush()
			out = append(out, string(r))
			lastClass = 0
			continue
		}
		c := classOf(r)
		if c == 0 {
			flush()
			out = append(out, string(r))
			lastClass = 0
			continue
		}
		// split on class change; also split each kanji (common baseline)
		if c == 2 {
			flush()
			out = append(out, string(r))
			lastClass = 2
			continue
		}
		if lastClass != 0 && c != lastClass {
			flush()
		}
		buf.WriteRune(r)
		lastClass = c
	}
	flush()
	return out
}
