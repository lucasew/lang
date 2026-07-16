package km

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"

// KhmerWordTokenizer ports tokenizers.km.KhmerWordTokenizer.
type KhmerWordTokenizer struct {
	*tokenizers.WordTokenizer
}

func NewKhmerWordTokenizer() *KhmerWordTokenizer {
	return &KhmerWordTokenizer{WordTokenizer: tokenizers.NewWordTokenizer()}
}

// khmerDelims are StringTokenizer delimiters from Java (returnDelims=true).
const khmerDelims = "\u17D4\u17D5\u0020\u00A0\u115f\u1160\u1680" +
	"\u2000\u2001\u2002\u2003\u2004\u2005\u2006\u2007" +
	"\u2008\u2009\u200A\u200B\u200c\u200d\u200e\u200f" +
	"\u2028\u2029\u202a\u202b\u202c\u202d\u202e\u202f" +
	"\u205F\u2060\u2061\u2062\u2063\u206A\u206b\u206c\u206d" +
	"\u206E\u206F\u3000\u3164\ufeff\uffa0\ufff9\ufffa\ufffb" +
	",.;()[]{}«»!?:\"'’‘„“”…\\/\t\n"

func (w *KhmerWordTokenizer) Tokenize(text string) []string {
	if text == "" {
		return nil
	}
	set := map[rune]bool{}
	for _, r := range khmerDelims {
		set[r] = true
	}
	var out []string
	var cur []rune
	flush := func() {
		if len(cur) > 0 {
			out = append(out, string(cur))
			cur = nil
		}
	}
	for _, r := range text {
		if set[r] {
			flush()
			out = append(out, string(r))
		} else {
			cur = append(cur, r)
		}
	}
	flush()
	return tokenizers.JoinEMailsAndUrls(out)
}
