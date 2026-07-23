package tokenizers

import "strings"

// ArabicWordTokenizer ports org.languagetool.tokenizers.ArabicWordTokenizer.
// Java: extends WordTokenizer; getTokenizingCharacters = super + "،؟؛-"
// (، = U+060C Arabic comma, ؟ = U+061F Arabic question mark, ؛ = U+061B Arabic
// semicolon, - = U+002D ASCII hyphen-minus). tokenize() is inherited from
// WordTokenizer unchanged (StringTokenizer keep-delims + joinEMailsAndUrls).
type ArabicWordTokenizer struct {
	// Cached super.getTokenizingCharacters() + "،؟؛-" (Java recomputes; result is constant).
	arTokenizingChars string
}

// NewArabicWordTokenizer ports ArabicWordTokenizer().
func NewArabicWordTokenizer() *ArabicWordTokenizer {
	return &ArabicWordTokenizer{
		arTokenizingChars: TokenizingCharacters() + "،؟؛-",
	}
}

// GetTokenizingCharacters ports ArabicWordTokenizer.getTokenizingCharacters.
func (w *ArabicWordTokenizer) GetTokenizingCharacters() string {
	return w.arTokenizingChars
}

// Tokenize ports the inherited WordTokenizer.tokenize using Arabic delims.
// Java: StringTokenizer(text, getTokenizingCharacters(), true) then joinEMailsAndUrls.
func (w *ArabicWordTokenizer) Tokenize(text string) []string {
	delims := w.arTokenizingChars
	out := make([]string, 0)
	var cur []rune
	flush := func() {
		if len(cur) > 0 {
			out = append(out, string(cur))
			cur = nil
		}
	}
	for _, r := range text {
		if strings.ContainsRune(delims, r) {
			flush()
			out = append(out, string(r))
		} else {
			cur = append(cur, r)
		}
	}
	flush()
	return JoinEMailsAndUrls(out)
}
