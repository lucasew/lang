package nl

import (
	"strings"
	"unicode/utf16"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// DutchWordTokenizer ports org.languagetool.tokenizers.nl.DutchWordTokenizer.
type DutchWordTokenizer struct {
	nlTokenizingChars string
}

var dutchQuotes = []string{"'", "`", "’", "‘", "´"}

func NewDutchWordTokenizer() *DutchWordTokenizer {
	chars := tokenizers.TokenizingCharacters() + "_"
	for _, q := range dutchQuotes {
		chars = strings.ReplaceAll(chars, q, "")
	}
	return &DutchWordTokenizer{nlTokenizingChars: chars}
}

func (w *DutchWordTokenizer) GetTokenizingCharacters() string {
	return w.nlTokenizingChars
}

func (w *DutchWordTokenizer) Tokenize(text string) []string {
	raw := splitKeepDelims(text, w.nlTokenizingChars)
	var l []string
	for _, token := range raw {
		origToken := token
		// Java: token.length() / substring on UTF-16 code units
		if tokenizers.UTF16Len(token) > 1 {
			if startsWithQuote(token) && endsWithQuote(token) && tokenizers.UTF16Len(token) > 2 {
				n := tokenizers.UTF16Len(token)
				l = append(l, utf16Sub(token, 0, 1))
				l = append(l, utf16Sub(token, 1, n-1))
				l = append(l, utf16Sub(token, n-1, n))
			} else if endsWithQuote(token) {
				cnt := 0
				for endsWithQuote(token) {
					n := tokenizers.UTF16Len(token)
					token = utf16Sub(token, 0, n-1)
					cnt++
				}
				l = append(l, token)
				origN := tokenizers.UTF16Len(origToken)
				for i := origN - cnt; i < origN; i++ {
					l = append(l, utf16Sub(origToken, i, i+1))
				}
			} else if startsWithQuote(token) {
				for startsWithQuote(token) {
					l = append(l, utf16Sub(token, 0, 1))
					token = utf16Sub(token, 1, tokenizers.UTF16Len(token))
				}
				l = append(l, token)
			} else {
				l = append(l, token)
			}
		} else {
			l = append(l, token)
		}
	}
	return tokenizers.JoinEMailsAndUrls(l)
}

// utf16Sub ports Java String.substring(from, to) with UTF-16 indices.
func utf16Sub(s string, from, to int) string {
	u := utf16.Encode([]rune(s))
	if from < 0 {
		from = 0
	}
	if to > len(u) {
		to = len(u)
	}
	if from >= to {
		return ""
	}
	return string(utf16.Decode(u[from:to]))
}

func startsWithQuote(token string) bool {
	for _, q := range dutchQuotes {
		if strings.HasPrefix(token, q) {
			return true
		}
	}
	return false
}

func endsWithQuote(token string) bool {
	for _, q := range dutchQuotes {
		if strings.HasSuffix(token, q) {
			return true
		}
	}
	return false
}

// splitKeepDelims is StringTokenizer(text, delims, true).
func splitKeepDelims(text, delims string) []string {
	if text == "" {
		return nil
	}
	var out []string
	var cur strings.Builder
	isDelim := func(r rune) bool {
		return strings.ContainsRune(delims, r)
	}
	for _, r := range text {
		if isDelim(r) {
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
