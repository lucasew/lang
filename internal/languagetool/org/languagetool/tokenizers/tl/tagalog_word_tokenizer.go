package tl

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// TagalogWordTokenizer ports org.languagetool.language.tokenizers.TagalogWordTokenizer.
// Java: extends WordTokenizer; getTokenizingCharacters = super + "-"
// (- = U+002D ASCII hyphen-minus). tokenize() is inherited from WordTokenizer
// unchanged (StringTokenizer keep-delims + joinEMailsAndUrls).
type TagalogWordTokenizer struct {
	// Cached super.getTokenizingCharacters() + "-" (Java recomputes; result is constant).
	tlTokenizingChars string
}

// NewTagalogWordTokenizer ports TagalogWordTokenizer().
func NewTagalogWordTokenizer() *TagalogWordTokenizer {
	return &TagalogWordTokenizer{
		tlTokenizingChars: tokenizers.TokenizingCharacters() + "-",
	}
}

// GetTokenizingCharacters ports TagalogWordTokenizer.getTokenizingCharacters.
// Java: return super.getTokenizingCharacters() + "-";
func (w *TagalogWordTokenizer) GetTokenizingCharacters() string {
	return w.tlTokenizingChars
}

// Tokenize ports the inherited WordTokenizer.tokenize using Tagalog delims.
// Java: StringTokenizer(text, getTokenizingCharacters(), true) then joinEMailsAndUrls.
func (w *TagalogWordTokenizer) Tokenize(text string) []string {
	raw := splitKeepDelims(text, w.tlTokenizingChars)
	return tokenizers.JoinEMailsAndUrls(raw)
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
