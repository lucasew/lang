package en

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// GoogleStyleWordTokenizer ports org.languagetool.rules.en.GoogleStyleWordTokenizer.
// Approximates Google ngram tokenization: hyphen splits; 'm/'re/'ve/'ll rejoin.
type GoogleStyleWordTokenizer struct {
	base *tokenizers.WordTokenizer
}

func NewGoogleStyleWordTokenizer() *GoogleStyleWordTokenizer {
	return &GoogleStyleWordTokenizer{base: tokenizers.NewWordTokenizer()}
}

func (w *GoogleStyleWordTokenizer) GetTokenizingCharacters() string {
	return tokenizers.TokenizingCharacters() + "-"
}

func (w *GoogleStyleWordTokenizer) Tokenize(text string) []string {
	// WordTokenizer uses fixed delims; re-tokenize with hyphen included via custom split.
	raw := splitKeepDelims(text, w.GetTokenizingCharacters())
	raw = tokenizers.JoinEMailsAndUrls(raw)
	var prev string
	var out []string
	for _, token := range raw {
		if prev == "'" {
			switch token {
			case "m":
				out[len(out)-1] = "'m"
			case "re":
				out[len(out)-1] = "'re"
			case "ve":
				out[len(out)-1] = "'ve"
			case "ll":
				out[len(out)-1] = "'ll"
			default:
				out = append(out, token)
			}
		} else {
			out = append(out, token)
		}
		prev = token
	}
	return out
}

func splitKeepDelims(text, delims string) []string {
	if text == "" {
		return nil
	}
	var out []string
	var cur []rune
	flush := func() {
		if len(cur) > 0 {
			out = append(out, string(cur))
			cur = nil
		}
	}
	for _, r := range text {
		if containsRune(delims, r) {
			flush()
			out = append(out, string(r))
		} else {
			cur = append(cur, r)
		}
	}
	flush()
	return out
}

func containsRune(s string, r rune) bool {
	for _, c := range s {
		if c == r {
			return true
		}
	}
	return false
}
