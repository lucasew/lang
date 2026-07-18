package ja

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// JapaneseWordTokenizer ports tokenizers.ja.JapaneseWordTokenizer.
// Full morphological analysis (Sen/Kuromoji) is deferred. Soft path uses
// longest-match over surfaces from ja-upstream-soft.xml (same multi-char
// tokens the soft rules expect) with single-char CJK fallback — a practical
// approximation of Java Sen dictionary segmentation for soft goldens.
type JapaneseWordTokenizer struct {
	// Segment optional custom segmenter (tests / full morph inject).
	Segment func(text string) []string
}

func NewJapaneseWordTokenizer() *JapaneseWordTokenizer { return &JapaneseWordTokenizer{} }

func (t *JapaneseWordTokenizer) Tokenize(text string) []string {
	if t != nil && t.Segment != nil {
		return t.Segment(text)
	}
	lex := tokenizers.SoftCJKLexiconForLang("ja")
	return tokenizers.SegmentCJKLongestMatch(text, lex)
}
