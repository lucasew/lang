package en

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tokenizers"
)

// GoogleStyleWordTokenizer ports org.languagetool.rules.en.GoogleStyleWordTokenizer.
//
// Tokenize sentences to tokens like Google does for its ngram index. Note: there
// doesn't seem to be official documentation about the way Google tokenizes there,
// so this is just an approximation.
//
// Java: extends WordTokenizer (@since 3.2).
type GoogleStyleWordTokenizer struct{}

// NewGoogleStyleWordTokenizer ports new GoogleStyleWordTokenizer().
func NewGoogleStyleWordTokenizer() *GoogleStyleWordTokenizer {
	return &GoogleStyleWordTokenizer{}
}

// GetTokenizingCharacters ports GoogleStyleWordTokenizer.getTokenizingCharacters.
// Java: return super.getTokenizingCharacters() + "-";
func (w *GoogleStyleWordTokenizer) GetTokenizingCharacters() string {
	// super.getTokenizingCharacters() is WordTokenizer.TOKENIZING_CHARACTERS
	return tokenizers.TokenizingCharacters() + "-"
}

// Tokenize ports GoogleStyleWordTokenizer.tokenize.
// Java:
//
//	List<String> tokens = super.tokenize(text);
//	String prev = null;
//	Stack<String> l = new Stack<>();
//	for (String token : tokens) {
//	  if ("'".equals(prev)) {
//	    // TODO: add more cases if needed:
//	    if (token.equals("m")) { l.pop(); l.push("'m"); }
//	    else if (token.equals("re")) { l.pop(); l.push("'re"); }
//	    else if (token.equals("ve")) { l.pop(); l.push("'ve"); }
//	    else if (token.equals("ll")) { l.pop(); l.push("'ll"); }
//	    else { l.push(token); }
//	  } else {
//	    l.push(token);
//	  }
//	  prev = token;
//	}
//	return l;
func (w *GoogleStyleWordTokenizer) Tokenize(text string) []string {
	tokens := w.superTokenize(text)
	// Java: String prev = null; "'".equals(null) is false. Go: empty never equals "'".
	prev := ""
	// Java: Stack<String> l = new Stack<>(); push/pop at end ≈ append/slice.
	l := make([]string, 0, len(tokens))
	for _, token := range tokens {
		if prev == "'" {
			// TODO: add more cases if needed:
			if token == "m" {
				l = l[:len(l)-1] // pop
				l = append(l, "'m")
			} else if token == "re" {
				l = l[:len(l)-1]
				l = append(l, "'re")
			} else if token == "ve" {
				l = l[:len(l)-1]
				l = append(l, "'ve")
			} else if token == "ll" {
				l = l[:len(l)-1]
				l = append(l, "'ll")
			} else {
				l = append(l, token)
			}
		} else {
			l = append(l, token)
		}
		prev = token
	}
	return l
}

// superTokenize ports WordTokenizer.tokenize with this type's getTokenizingCharacters.
// Java WordTokenizer.tokenize:
//
//	StringTokenizer st = new StringTokenizer(text, getTokenizingCharacters(), true);
//	while (st.hasMoreElements()) { l.add(st.nextToken()); }
//	return joinEMailsAndUrls(l);
func (w *GoogleStyleWordTokenizer) superTokenize(text string) []string {
	raw := stringTokenizerKeepDelims(text, w.GetTokenizingCharacters())
	return tokenizers.JoinEMailsAndUrls(raw)
}

// stringTokenizerKeepDelims ports Java StringTokenizer(text, delims, true):
// non-delimiter runs become tokens; each delimiter char is its own token.
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
