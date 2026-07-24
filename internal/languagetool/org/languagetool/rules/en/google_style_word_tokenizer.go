package en

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// GoogleStyleWordTokenizer ports org.languagetool.rules.en.GoogleStyleWordTokenizer.
// Approximates Google ngram tokenization: hyphen splits; 'm/'re/'ve/'ll rejoin.
// Java: extends WordTokenizer.
type GoogleStyleWordTokenizer struct{}

func NewGoogleStyleWordTokenizer() *GoogleStyleWordTokenizer {
	return &GoogleStyleWordTokenizer{}
}

// GetTokenizingCharacters ports GoogleStyleWordTokenizer.getTokenizingCharacters.
// Java: return super.getTokenizingCharacters() + "-";
func (w *GoogleStyleWordTokenizer) GetTokenizingCharacters() string {
	return tokenizers.TokenizingCharacters() + "-"
}

// Tokenize ports GoogleStyleWordTokenizer.tokenize.
// Java: super.tokenize(text) then Stack re-glue of 'm/'re/'ve/'ll after apostrophe.
func (w *GoogleStyleWordTokenizer) Tokenize(text string) []string {
	// super.tokenize: StringTokenizer(text, getTokenizingCharacters(), true) + joinEMailsAndUrls
	tokens := stringTokenizerKeepDelims(text, w.GetTokenizingCharacters())
	tokens = tokenizers.JoinEMailsAndUrls(tokens)

	var prev string // Java: String prev = null; "" never equals "'"
	// Java: Stack<String> l = new Stack<>();
	l := make([]string, 0, len(tokens))
	for _, token := range tokens {
		if prev == "'" {
			// TODO (Java): add more cases if needed
			switch token {
			case "m":
				l = l[:len(l)-1] // pop apostrophe
				l = append(l, "'m")
			case "re":
				l = l[:len(l)-1]
				l = append(l, "'re")
			case "ve":
				l = l[:len(l)-1]
				l = append(l, "'ve")
			case "ll":
				l = l[:len(l)-1]
				l = append(l, "'ll")
			default:
				l = append(l, token)
			}
		} else {
			l = append(l, token)
		}
		prev = token
	}
	return l
}

// stringTokenizerKeepDelims ports Java StringTokenizer(text, delims, true).
func stringTokenizerKeepDelims(text, delims string) []string {
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
	return out
}
