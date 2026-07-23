package ml

import "strings"

// MalayalamWordTokenizer ports org.languagetool.tokenizers.ml.MalayalamWordTokenizer.
// Java: implements Tokenizer directly (does NOT extend WordTokenizer).
// tokenize() uses a hardcoded StringTokenizer delimiter string with
// returnDelims=true and returns the raw token list — does NOT call
// joinEMailsAndUrls (contrast WordTokenizer / KhmerWordTokenizer).
type MalayalamWordTokenizer struct{}

// NewMalayalamWordTokenizer ports MalayalamWordTokenizer().
func NewMalayalamWordTokenizer() *MalayalamWordTokenizer {
	return &MalayalamWordTokenizer{}
}

// malayalamDelims is the exact Java StringTokenizer delimiter string from
// MalayalamWordTokenizer.tokenize (returnDelims=true). Character-for-character
// match of the Java literal:
//
//	"\u0020\u00A0\u115f\u1160\u1680" + ",.;()[]{}!?:\"'’‘„“”…\\/\t\n"
//
// Smaller set than core WordTokenizer (omits CR, VT, en-dash, '=', «», many
// Unicode spaces, etc.).
const malayalamDelims = "\u0020\u00A0\u115f\u1160\u1680" +
	",.;()[]{}!?:\"'’‘„“”…\\/\t\n"

// Tokenize ports MalayalamWordTokenizer.tokenize.
// Java: StringTokenizer(text, <hardcoded delims>, true); return tokens as-is.
func (w *MalayalamWordTokenizer) Tokenize(text string) []string {
	return splitKeepDelims(text, malayalamDelims)
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
