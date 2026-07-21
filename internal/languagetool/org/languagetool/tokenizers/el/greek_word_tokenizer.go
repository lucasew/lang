package el

import (
	"strings"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// GreekWordTokenizer ports tokenizers.el.GreekWordTokenizer.
// Java: GreekWordTokenizerImpl (JFlex) then WordTokenizer.joinEMailsAndUrls.
// Delim set and special token "ό,τι" come from GreekWordTokenizerImpl.jflex.
type GreekWordTokenizer struct{}

func NewGreekWordTokenizer() *GreekWordTokenizer { return &GreekWordTokenizer{} }

// greekDelim is the JFlex Delim production (single-code-point delimiters).
// Each Delim char is its own token; maximal non-Delim runs are word tokens.
// Special multi-char token "ό,τι" (contains comma) is matched first.
var greekDelim = buildGreekDelimSet()

func buildGreekDelimSet() map[rune]struct{} {
	// From GreekWordTokenizerImpl.jflex Delim =
	chars := "" +
		"\u0020\u00A0\u115f\u1160\u1680" +
		"\u2000\u2001\u2002\u2003\u2004\u2005\u2006\u2007" +
		"\u2008\u2009\u200A\u200B\u200c\u200d\u200e\u200f" +
		"\u2028\u2029\u202a\u202b\u202c\u202d\u202e\u202f" +
		"\u205F\u2060\u2061\u2062\u2063\u206A\u206b\u206c\u206d" +
		"\u206E\u206F\u3000\u3164\ufeff\uffa0\ufff9\ufffa\ufffb" +
		",.;()[]{}!:\"'" +
		"·" + // Greek Ano Teleia
		"’‘„“”…«»\\/\t\n"
	m := make(map[rune]struct{}, len(chars))
	for _, r := range chars {
		m[r] = struct{}{}
	}
	return m
}

const greekSpecialOti = "ό,τι"

func (t *GreekWordTokenizer) Tokenize(text string) []string {
	if text == "" {
		return nil
	}
	raw := greekJflexTokenize(text)
	return tokenizers.JoinEMailsAndUrls(raw)
}

// greekJflexTokenize ports the JFlex Word production:
// Word = "ό,τι" | (!Delim)* | Delim
// (maximal match: special first, else non-delim run, else single delim).
func greekJflexTokenize(text string) []string {
	var out []string
	i := 0
	for i < len(text) {
		// Special multi-char token (comma is otherwise a Delim).
		if strings.HasPrefix(text[i:], greekSpecialOti) {
			out = append(out, greekSpecialOti)
			i += len(greekSpecialOti)
			continue
		}
		r, size := utf8.DecodeRuneInString(text[i:])
		if r == utf8.RuneError && size == 1 {
			out = append(out, text[i:i+1])
			i++
			continue
		}
		if _, isDelim := greekDelim[r]; isDelim {
			out = append(out, string(r))
			i += size
			continue
		}
		// Maximal non-Delim run.
		start := i
		i += size
		for i < len(text) {
			if strings.HasPrefix(text[i:], greekSpecialOti) {
				// "ό,τι" starts a new token; stop the non-delim run.
				break
			}
			r2, sz2 := utf8.DecodeRuneInString(text[i:])
			if _, isDelim := greekDelim[r2]; isDelim {
				break
			}
			i += sz2
		}
		out = append(out, text[start:i])
	}
	return out
}
