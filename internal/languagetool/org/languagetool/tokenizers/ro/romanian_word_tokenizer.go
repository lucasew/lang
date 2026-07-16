package ro

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// RomanianWordTokenizer ports org.languagetool.tokenizers.ro.RomanianWordTokenizer.
type RomanianWordTokenizer struct{}

func NewRomanianWordTokenizer() *RomanianWordTokenizer {
	return &RomanianWordTokenizer{}
}

// Explicit delimiter set from Java RomanianWordTokenizer.tokenize.
const roDelims = "\u0020\u00A0\u115f\u1160\u1680" +
	"\u2000\u2001\u2002\u2003\u2004\u2005\u2006\u2007" +
	"\u2008\u2009\u200A\u200B\u200c\u200d\u200e\u200f" +
	"\u2028\u2029\u202a\u202b\u202c\u202d\u202e\u202f" +
	"\u205F\u2060\u2061\u2062\u2063\u206A\u206b\u206c\u206d" +
	"\u206E\u206F\u3000\u3164\ufeff\uffa0\ufff9\ufffa\ufffb" +
	",.;()[]{}!?:\"'’‘„“”…\\/\t\n\r«»<>%°" + "-|="

func (w *RomanianWordTokenizer) Tokenize(text string) []string {
	raw := splitKeepDelims(text, roDelims)
	return tokenizers.JoinEMailsAndUrls(raw)
}

func splitKeepDelims(text, delims string) []string {
	if text == "" {
		return nil
	}
	var out []string
	var cur strings.Builder
	for _, r := range text {
		if strings.ContainsRune(delims, r) {
			if cur.Len() > 0 {
				out = append(out, cur.String())
				cur.Reset()
			}
			out = append(out, string(r))
		} else {
			cur.WriteRune(r)
		}
	}
	if cur.Len() > 0 {
		out = append(out, cur.String())
	}
	return out
}
