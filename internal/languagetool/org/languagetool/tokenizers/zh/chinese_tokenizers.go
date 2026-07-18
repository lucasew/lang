package zh

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// ChineseWordTokenizer ports tokenizers.zh.ChineseWordTokenizer.
//
// Java uses HanLP and returns each Term as Term.toString() → "surface/pos".
// Full HanLP is deferred; soft path segments with SoftCJKLexicon longest-match
// then encodes "surface/pos" so ChineseTagger.asAnalyzedToken can decode like Java.
type ChineseWordTokenizer struct {
	// Segment optional custom segmenter (surfaces only; POS still soft-guessed).
	Segment func(text string) []string
}

func NewChineseWordTokenizer() *ChineseWordTokenizer { return &ChineseWordTokenizer{} }

func (t *ChineseWordTokenizer) Tokenize(text string) []string {
	if t != nil && t.Segment != nil {
		return encodeChineseTerms(t.Segment(text))
	}
	lex := tokenizers.SoftCJKLexiconForLang("zh")
	return encodeChineseTerms(tokenizers.SegmentCJKLongestMatch(text, lex))
}

// encodeChineseTerms maps surfaces to Java HanLP Term.toString form "surface/pos".
func encodeChineseTerms(surfaces []string) []string {
	if len(surfaces) == 0 {
		return nil
	}
	out := make([]string, 0, len(surfaces))
	for _, s := range surfaces {
		if s == "" {
			continue
		}
		out = append(out, s+"/"+softGuessHanLPStylePOS(s))
	}
	return out
}

// softGuessHanLPStylePOS assigns a coarse HanLP-style short tag without HanLP.
// Unknown CJK defaults to "x" (untagged soft path) so surface-only goldens stay stable.
func softGuessHanLPStylePOS(s string) string {
	if s == "" {
		return "x"
	}
	// Pure punctuation / symbols → w (HanLP punctuation)
	allPunct := true
	allDigit := true
	allLatin := true
	hasHan := false
	for _, r := range s {
		if unicode.Is(unicode.Han, r) {
			hasHan = true
		}
		if !unicode.IsPunct(r) && !unicode.IsSymbol(r) && !unicode.IsSpace(r) {
			allPunct = false
		}
		if !unicode.IsDigit(r) {
			allDigit = false
		}
		if r > 127 || !(unicode.IsLetter(r) || unicode.IsDigit(r)) {
			if r <= 127 && (unicode.IsLetter(r) || unicode.IsDigit(r)) {
				// ok
			} else if r > 127 {
				allLatin = false
			}
		}
		if r > 127 {
			allLatin = false
		}
	}
	// recompute allLatin simply
	allLatin = true
	for _, r := range s {
		if r > 127 || !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '\'') {
			allLatin = false
			break
		}
	}
	if allPunct {
		return "w"
	}
	if allDigit {
		return "m"
	}
	if allLatin && !hasHan {
		return "nx"
	}
	// Soft: leave open-class CJK as "x" so pattern soft POS matching (empty tag)
	// still treats them as untagged words until HanLP is wired.
	return "x"
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
