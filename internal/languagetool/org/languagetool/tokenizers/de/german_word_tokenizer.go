package de

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// GermanWordTokenizer ports org.languagetool.tokenizers.de.GermanWordTokenizer.
// Adds underscore and single low-9 quotation mark ‚ as tokenizing characters.
type GermanWordTokenizer struct {
	delims string
}

func NewGermanWordTokenizer() *GermanWordTokenizer {
	return &GermanWordTokenizer{
		delims: tokenizers.TokenizingCharacters() + "_‚",
	}
}

func (w *GermanWordTokenizer) GetTokenizingCharacters() string { return w.delims }

func (w *GermanWordTokenizer) Tokenize(text string) []string {
	return tokenizers.JoinEMailsAndUrls(splitKeepDelims(text, w.delims))
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
