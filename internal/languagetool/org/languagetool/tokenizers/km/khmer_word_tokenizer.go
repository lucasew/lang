package km

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// KhmerWordTokenizer ports org.languagetool.tokenizers.km.KhmerWordTokenizer.
// Java: extends WordTokenizer; overrides tokenize() only with a hardcoded
// StringTokenizer delimiter string (does NOT call getTokenizingCharacters()).
// Inherited getTokenizingCharacters() remains the base WordTokenizer set.
// tokenize ends with joinEMailsAndUrls(tokens).
type KhmerWordTokenizer struct {
	*tokenizers.WordTokenizer
}

// NewKhmerWordTokenizer ports KhmerWordTokenizer().
func NewKhmerWordTokenizer() *KhmerWordTokenizer {
	return &KhmerWordTokenizer{WordTokenizer: tokenizers.NewWordTokenizer()}
}

// khmerDelims is the exact Java StringTokenizer delimiter string from
// KhmerWordTokenizer.tokenize (returnDelims=true). Character-for-character
// match of the Java literal (includes Khmer signs U+17D4 ។ and U+17D5 ៕;
// omits many base WordTokenizer delims such as ASCII hyphen-minus, CR, VT).
const khmerDelims = "\u17D4\u17D5\u0020\u00A0\u115f\u1160\u1680" +
	"\u2000\u2001\u2002\u2003\u2004\u2005\u2006\u2007" +
	"\u2008\u2009\u200A\u200B\u200c\u200d\u200e\u200f" +
	"\u2028\u2029\u202a\u202b\u202c\u202d\u202e\u202f" +
	"\u205F\u2060\u2061\u2062\u2063\u206A\u206b\u206c\u206d" +
	"\u206E\u206F\u3000\u3164\ufeff\uffa0\ufff9\ufffa\ufffb" +
	",.;()[]{}«»!?:\"'’‘„“”…\\/\t\n"

// Tokenize ports KhmerWordTokenizer.tokenize.
// Java: StringTokenizer(text, <hardcoded delims>, true) then joinEMailsAndUrls.
func (w *KhmerWordTokenizer) Tokenize(text string) []string {
	raw := splitKeepDelims(text, khmerDelims)
	return tokenizers.JoinEMailsAndUrls(raw)
}

// splitKeepDelims is StringTokenizer(text, delims, true).
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
