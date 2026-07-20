package be

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// BelarusianWordTokenizer ports org.languagetool.tokenizers.be.BelarusianWordTokenizer.
// Apostrophes (' ’ ʼ) are part of the word.
type BelarusianWordTokenizer struct {
	tokenizingCharacters string
}

func NewBelarusianWordTokenizer() *BelarusianWordTokenizer {
	chars := tokenizers.TokenizingCharacters()
	chars = strings.ReplaceAll(chars, "'", "")
	chars = strings.ReplaceAll(chars, "’", "")
	chars = strings.ReplaceAll(chars, "ʼ", "")
	return &BelarusianWordTokenizer{tokenizingCharacters: chars}
}

func (w *BelarusianWordTokenizer) GetTokenizingCharacters() string {
	return w.tokenizingCharacters
}

func (w *BelarusianWordTokenizer) Tokenize(text string) []string {
	raw := splitKeepDelims(text, w.tokenizingCharacters)
	joined := tokenizers.JoinEMailsAndUrls(raw)
	var out []string
	for _, token := range joined {
		if tokenizers.UTF16Len(token) > 1 {
			out = append(out, strings.ReplaceAll(token, "’", "'"))
		} else {
			out = append(out, token)
		}
	}
	return out
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
