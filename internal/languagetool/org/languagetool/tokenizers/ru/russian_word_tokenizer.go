package ru

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// RussianWordTokenizer ports org.languagetool.tokenizers.ru.RussianWordTokenizer.
// Java: extends WordTokenizer; getTokenizingCharacters = super + "'."; tokenize
// protects б/у and б/н, guards space-dot patterns so trailing " ." becomes space + ".",
// StringTokenizer(delims, true), wordsToAdd identity, joinEMailsAndUrls.
type RussianWordTokenizer struct {
	delims string
}

// NewRussianWordTokenizer ports RussianWordTokenizer().
func NewRussianWordTokenizer() *RussianWordTokenizer {
	return &RussianWordTokenizer{
		// Java: super.getTokenizingCharacters() + "'."
		delims: tokenizers.TokenizingCharacters() + "'.",
	}
}

// GetTokenizingCharacters ports RussianWordTokenizer.getTokenizingCharacters.
func (w *RussianWordTokenizer) GetTokenizingCharacters() string { return w.delims }

// Tokenize ports RussianWordTokenizer.tokenize.
func (w *RussianWordTokenizer) Tokenize(text string) []string {
	// Java: auxText chain of replace (String.replace = all occurrences).
	aux := text
	aux = strings.ReplaceAll(aux, "б/у", "\u0001\u0001SOCR_BU\u0001\u0001")
	aux = strings.ReplaceAll(aux, "б/н", "\u0001\u0001SOCR_BN\u0001\u0001")
	aux = strings.ReplaceAll(aux, " .. ", "\u0001\u0001SP_DDOT_SP\u0001\u0001")
	aux = strings.ReplaceAll(aux, " . ", "\u0001\u0001SP_DOT_SP\u0001\u0001")
	aux = strings.ReplaceAll(aux, " .", " \u0001\u0001SP_DOT\u0001\u0001")
	// Restore " .. " / " . " before StringTokenizer so only SP_DOT stays protected.
	aux = strings.ReplaceAll(aux, "\u0001\u0001SP_DDOT_SP\u0001\u0001", " .. ")
	aux = strings.ReplaceAll(aux, "\u0001\u0001SP_DOT_SP\u0001\u0001", " . ")

	// Java: StringTokenizer(auxText, getTokenizingCharacters(), true)
	raw := splitKeepDelims(aux, w.delims)
	var l []string
	for _, s := range raw {
		s = strings.ReplaceAll(s, "\u0001\u0001SOCR_BU\u0001\u0001", "б/у")
		s = strings.ReplaceAll(s, "\u0001\u0001SOCR_BN\u0001\u0001", "б/н")
		s = strings.ReplaceAll(s, "\u0001\u0001SP_DOT\u0001\u0001", ".")
		l = append(l, wordsToAdd(s)...)
	}
	return tokenizers.JoinEMailsAndUrls(l)
}

// wordsToAdd ports RussianWordTokenizer.wordsToAdd (identity).
func wordsToAdd(s string) []string {
	return []string{s}
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
