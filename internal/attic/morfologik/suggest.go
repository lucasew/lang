package morfologik

import (
	"strings"
	"unicode"
)

// SuggestEdits returns dictionary words within one Damerau-Levenshtein edit of word
// by generating candidates and testing Contains (works for large CFSA2 dicts).
// max caps the result size (0 → 8).
func (d *Dictionary) SuggestEdits(word string, max int) []string {
	if d == nil || word == "" {
		return nil
	}
	if max <= 0 {
		max = 8
	}
	low := strings.ToLower(word)
	// Prefer lowercase probe; keep original if all-caps suggestions needed later.
	cands := edit1Candidates(low)
	var out []string
	seen := map[string]struct{}{}
	for _, c := range cands {
		if c == low {
			continue
		}
		if _, ok := seen[c]; ok {
			continue
		}
		if !d.Contains(c) {
			continue
		}
		seen[c] = struct{}{}
		// Preserve capitalization hint: if input was Capitalized, title-case suggestion
		sug := c
		if isTitleCase(word) {
			sug = titleCase(c)
		} else if isAllUpper(word) {
			sug = strings.ToUpper(c)
		}
		out = append(out, sug)
		if len(out) >= max {
			break
		}
	}
	return out
}

func isTitleCase(s string) bool {
	r := []rune(s)
	if len(r) < 2 {
		return false
	}
	return unicode.IsUpper(r[0]) && strings.ToLower(s[1:]) == s[1:]
}

func isAllUpper(s string) bool {
	has := false
	for _, r := range s {
		if unicode.IsLetter(r) {
			has = true
			if !unicode.IsUpper(r) {
				return false
			}
		}
	}
	return has
}

func titleCase(s string) string {
	r := []rune(s)
	if len(r) == 0 {
		return s
	}
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

// edit1Candidates generates lowercase distance-1 candidates (insert/delete/replace/transpose).
func edit1Candidates(word string) []string {
	letters := "abcdefghijklmnopqrstuvwxyz'"
	w := []rune(word)
	n := len(w)
	// estimate capacity
	out := make([]string, 0, n*27+n*26+n)
	// deletes
	for i := 0; i < n; i++ {
		out = append(out, string(append(append([]rune{}, w[:i]...), w[i+1:]...)))
	}
	// transposes
	for i := 0; i < n-1; i++ {
		rw := append([]rune{}, w...)
		rw[i], rw[i+1] = rw[i+1], rw[i]
		out = append(out, string(rw))
	}
	// replaces
	for i := 0; i < n; i++ {
		for _, c := range letters {
			if w[i] == c {
				continue
			}
			rw := append([]rune{}, w...)
			rw[i] = c
			out = append(out, string(rw))
		}
	}
	// inserts
	for i := 0; i <= n; i++ {
		for _, c := range letters {
			rw := make([]rune, 0, n+1)
			rw = append(rw, w[:i]...)
			rw = append(rw, c)
			rw = append(rw, w[i:]...)
			out = append(out, string(rw))
		}
	}
	return out
}
