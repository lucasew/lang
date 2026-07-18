package zh

import (
	"strings"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// ChineseWordTokenizer ports tokenizers.zh.ChineseWordTokenizer.
// Full HanLP segmentation is deferred. Soft path uses longest-match over
// surfaces from zh-upstream-soft.xml with single-char Han fallback.
type ChineseWordTokenizer struct {
	// Segment optional custom segmenter.
	Segment func(text string) []string
}

func NewChineseWordTokenizer() *ChineseWordTokenizer { return &ChineseWordTokenizer{} }

func (t *ChineseWordTokenizer) Tokenize(text string) []string {
	if t != nil && t.Segment != nil {
		return t.Segment(text)
	}
	lex := tokenizers.SoftCJKLexiconForLang("zh")
	return tokenizers.SegmentCJKLongestMatch(text, lex)
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
