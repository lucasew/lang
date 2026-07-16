package nl

import (
	"strings"

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
		if len([]rune(token)) > 1 {
			if startsWithQuote(token) && endsWithQuote(token) && len([]rune(token)) > 2 {
				rs := []rune(token)
				l = append(l, string(rs[0]))
				l = append(l, string(rs[1:len(rs)-1]))
				l = append(l, string(rs[len(rs)-1]))
			} else if endsWithQuote(token) {
				cnt := 0
				for endsWithQuote(token) {
					rs := []rune(token)
					token = string(rs[:len(rs)-1])
					cnt++
				}
				l = append(l, token)
				origRunes := []rune(origToken)
				for i := len(origRunes) - cnt; i < len(origRunes); i++ {
					l = append(l, string(origRunes[i]))
				}
			} else if startsWithQuote(token) {
				for startsWithQuote(token) {
					rs := []rune(token)
					l = append(l, string(rs[0]))
					token = string(rs[1:])
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
