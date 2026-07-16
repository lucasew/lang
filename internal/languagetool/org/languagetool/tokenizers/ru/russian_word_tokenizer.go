package ru

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// RussianWordTokenizer ports org.languagetool.tokenizers.ru.RussianWordTokenizer.
type RussianWordTokenizer struct {
	delims string
}

func NewRussianWordTokenizer() *RussianWordTokenizer {
	return &RussianWordTokenizer{
		delims: tokenizers.TokenizingCharacters() + "'.",
	}
}

func (w *RussianWordTokenizer) GetTokenizingCharacters() string { return w.delims }

func (w *RussianWordTokenizer) Tokenize(text string) []string {
	aux := text
	aux = strings.ReplaceAll(aux, "б/у", "\u0001\u0001SOCR_BU\u0001\u0001")
	aux = strings.ReplaceAll(aux, "б/н", "\u0001\u0001SOCR_BN\u0001\u0001")
	aux = strings.ReplaceAll(aux, " .. ", "\u0001\u0001SP_DDOT_SP\u0001\u0001")
	aux = strings.ReplaceAll(aux, " . ", "\u0001\u0001SP_DOT_SP\u0001\u0001")
	aux = strings.ReplaceAll(aux, " .", " \u0001\u0001SP_DOT\u0001\u0001")
	aux = strings.ReplaceAll(aux, "\u0001\u0001SP_DDOT_SP\u0001\u0001", " .. ")
	aux = strings.ReplaceAll(aux, "\u0001\u0001SP_DOT_SP\u0001\u0001", " . ")

	raw := splitKeepDelims(aux, w.delims)
	var l []string
	for _, s := range raw {
		s = strings.ReplaceAll(s, "\u0001\u0001SOCR_BU\u0001\u0001", "б/у")
		s = strings.ReplaceAll(s, "\u0001\u0001SOCR_BN\u0001\u0001", "б/н")
		s = strings.ReplaceAll(s, "\u0001\u0001SP_DOT\u0001\u0001", ".")
		l = append(l, s)
	}
	return tokenizers.JoinEMailsAndUrls(l)
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
