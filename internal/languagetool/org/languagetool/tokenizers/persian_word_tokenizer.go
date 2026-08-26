package tokenizers

import "strings"

// PersianWordTokenizer ports org.languagetool.tokenizers.PersianWordTokenizer.
// Java: extends WordTokenizer; getTokenizingCharacters = super + "،؟؛"
// (، = U+060C Arabic comma, ؟ = U+061F Arabic question mark, ؛ = U+061B Arabic
// semicolon). Unlike ArabicWordTokenizer, does NOT append ASCII hyphen-minus.
// tokenize() is inherited from WordTokenizer unchanged (StringTokenizer
// keep-delims + joinEMailsAndUrls).
type PersianWordTokenizer struct {
	// Cached super.getTokenizingCharacters() + "،؟؛" (Java recomputes; result is constant).
	faTokenizingChars string
}

// NewPersianWordTokenizer ports PersianWordTokenizer().
func NewPersianWordTokenizer() *PersianWordTokenizer {
	return &PersianWordTokenizer{
		faTokenizingChars: TokenizingCharacters() + "،؟؛",
	}
}

// GetTokenizingCharacters ports PersianWordTokenizer.getTokenizingCharacters.
func (w *PersianWordTokenizer) GetTokenizingCharacters() string {
	return w.faTokenizingChars
}

// Tokenize ports the inherited WordTokenizer.tokenize using Persian delims.
// Java: StringTokenizer(text, getTokenizingCharacters(), true) then joinEMailsAndUrls.
func (w *PersianWordTokenizer) Tokenize(text string) []string {
	delims := w.faTokenizingChars
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
