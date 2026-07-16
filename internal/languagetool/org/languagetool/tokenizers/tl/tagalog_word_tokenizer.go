package tl

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// TagalogWordTokenizer ports org.languagetool.language.tokenizers.TagalogWordTokenizer.
// Adds hyphen as an additional tokenizing character.
type TagalogWordTokenizer struct {
	delims string
}

func NewTagalogWordTokenizer() *TagalogWordTokenizer {
	return &TagalogWordTokenizer{
		delims: tokenizers.TokenizingCharacters() + "-",
	}
}

func (w *TagalogWordTokenizer) GetTokenizingCharacters() string {
	return w.delims
}

func (w *TagalogWordTokenizer) Tokenize(text string) []string {
	raw := splitKeepDelims(text, w.delims)
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
