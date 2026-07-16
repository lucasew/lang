package commandline

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"unicode"
)

// IsKnownWordFunc reports whether a token is known in the language dictionary.
type IsKnownWordFunc func(token string) bool

// CollectUnknownWords tokenizes text and returns sorted unique unknown tokens
// (letter/digit sequences only; skips pure punctuation).
func CollectUnknownWords(text string, isKnown IsKnownWordFunc) []string {
	if isKnown == nil {
		return nil
	}
	seen := map[string]struct{}{}
	var out []string
	for _, tok := range simpleTokenize(text) {
		if !isWordToken(tok) {
			continue
		}
		if isKnown(tok) {
			continue
		}
		// also try lower
		if isKnown(strings.ToLower(tok)) {
			continue
		}
		if _, ok := seen[tok]; ok {
			continue
		}
		seen[tok] = struct{}{}
		out = append(out, tok)
	}
	sort.Strings(out)
	return out
}

func isWordToken(s string) bool {
	has := false
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			has = true
			continue
		}
		// allow internal apostrophe/hyphen
		if r == '\'' || r == '-' || r == '’' {
			continue
		}
		return false
	}
	return has
}

func simpleTokenize(text string) []string {
	var out []string
	var cur strings.Builder
	flush := func() {
		if cur.Len() > 0 {
			out = append(out, cur.String())
			cur.Reset()
		}
	}
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '\'' || r == '-' || r == '’' {
			cur.WriteRune(r)
		} else {
			flush()
			if !unicode.IsSpace(r) {
				out = append(out, string(r))
			}
		}
	}
	flush()
	return out
}

// PrintUnknownWords writes "Unknown words: a, b" line.
func PrintUnknownWords(w io.Writer, words []string) {
	if w == nil || len(words) == 0 {
		return
	}
	fmt.Fprintf(w, "Unknown words: %s\n", strings.Join(words, ", "))
}
