package crh

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// CrimeanTatarWordTokenizer ports org.languagetool.tokenizers.crh.CrimeanTatarWordTokenizer.
type CrimeanTatarWordTokenizer struct {
	delims string
}

func NewCrimeanTatarWordTokenizer() *CrimeanTatarWordTokenizer {
	return &CrimeanTatarWordTokenizer{
		delims: tokenizers.TokenizingCharacters() + "–", // n-dash
	}
}

func (w *CrimeanTatarWordTokenizer) GetTokenizingCharacters() string {
	return w.delims
}

func (w *CrimeanTatarWordTokenizer) Tokenize(text string) []string {
	raw := splitKeepDelims(text, w.delims)
	var l []string
	for _, token := range raw {
		if len(token) > 1 && strings.HasSuffix(token, "-") {
			l = append(l, token[:len(token)-1], "-")
		} else {
			l = append(l, token)
		}
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
