package de

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// GermanWordTokenizer ports org.languagetool.tokenizers.de.GermanWordTokenizer.
// Java: extends WordTokenizer; getTokenizingCharacters = super + "_‚"
// (‚ = U+201A single low-9 quotation mark, not a comma). tokenize() is
// inherited from WordTokenizer unchanged (StringTokenizer keep-delims +
// joinEMailsAndUrls; no DE-only wordsToAdd/currency path).
type GermanWordTokenizer struct {
	// Java: private final String deTokenizingChars = super.getTokenizingCharacters() + "_‚";
	deTokenizingChars string
}

// NewGermanWordTokenizer ports GermanWordTokenizer().
func NewGermanWordTokenizer() *GermanWordTokenizer {
	return &GermanWordTokenizer{
		deTokenizingChars: tokenizers.TokenizingCharacters() + "_‚",
	}
}

// GetTokenizingCharacters ports GermanWordTokenizer.getTokenizingCharacters.
func (w *GermanWordTokenizer) GetTokenizingCharacters() string {
	return w.deTokenizingChars
}

// Tokenize ports the inherited WordTokenizer.tokenize using German delims.
// Java: StringTokenizer(text, getTokenizingCharacters(), true) then joinEMailsAndUrls.
func (w *GermanWordTokenizer) Tokenize(text string) []string {
	raw := splitKeepDelims(text, w.deTokenizingChars)
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
