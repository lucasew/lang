package morfologik

import (
	"strings"
	"unicode"
)

// FREQ_RANGES ports morfologik Speller.FREQ_RANGES ('Z'-'A'+1 = 26).
const FreqRanges = 26

// SuggestEdits returns dictionary words within one Damerau-Levenshtein edit of word
// by generating candidates and testing Contains (works for large CFSA2 dicts).
// max caps the result size (0 → 8).
func (d *Dictionary) SuggestEdits(word string, max int) []string {
	return d.SuggestEditsMax(word, max, 1)
}

// SuggestEditsMax ports Speller suggestions with maxEditDistance (1, 2, or 3).
// Distance-1 uses full edit1 candidate set; distance 2/3 expands via successive edit1
// on out-of-dict neighbors (bounded for large dictionaries).
func (d *Dictionary) SuggestEditsMax(word string, maxResults, maxEdit int) []string {
	if d == nil || word == "" {
		return nil
	}
	if maxResults <= 0 {
		maxResults = 8
	}
	if maxEdit < 1 {
		maxEdit = 1
	}
	if maxEdit > 3 {
		maxEdit = 3
	}
	low := strings.ToLower(word)
	seen := map[string]struct{}{}
	var out []string

	addIfKnown := func(c string) bool {
		if c == "" || c == low {
			return false
		}
		if _, ok := seen[c]; ok {
			return false
		}
		if !d.Contains(c) {
			return false
		}
		seen[c] = struct{}{}
		sug := c
		if isTitleCase(word) {
			sug = titleCase(c)
		} else if isAllUpper(word) {
			sug = strings.ToUpper(c)
		}
		out = append(out, sug)
		return true
	}

	// distance 1
	cands1 := edit1Candidates(low)
	for _, c := range cands1 {
		if addIfKnown(c) && len(out) >= maxResults {
			return out
		}
	}
	if maxEdit < 2 || len(out) >= maxResults {
		return out
	}

	// distance 2: edit1 of misspelled edit1 neighbors (capped)
	const maxNeighbors = 200
	n := 0
	for _, c := range cands1 {
		if d.Contains(c) {
			continue
		}
		n++
		if n > maxNeighbors {
			break
		}
		for _, c2 := range edit1Candidates(c) {
			if addIfKnown(c2) && len(out) >= maxResults {
				return out
			}
		}
	}
	if maxEdit < 3 || len(out) >= maxResults {
		return out
	}

	// distance 3: one more expansion round on a smaller frontier of distance-2 misspellings
	// Re-run limited edit1×2 style is expensive; use second hop from first batch of cands1 only.
	n = 0
	for _, c := range cands1 {
		if d.Contains(c) {
			continue
		}
		n++
		if n > 40 {
			break
		}
		inner := 0
		for _, c2 := range edit1Candidates(c) {
			if d.Contains(c2) {
				continue
			}
			inner++
			if inner > 30 {
				break
			}
			for _, c3 := range edit1Candidates(c2) {
				if addIfKnown(c3) && len(out) >= maxResults {
					return out
				}
			}
		}
	}
	return out
}

// WeightedEditSuggestions returns suggestions with Java-like weights:
// distance*FREQ_RANGES + FREQ_RANGES - frequency - 1 (lower is better).
func (d *Dictionary) WeightedEditSuggestions(word string, maxResults, maxEdit int) []struct {
	Word   string
	Weight int
} {
	sugs := d.SuggestEditsMax(word, maxResults, maxEdit)
	if len(sugs) == 0 {
		return nil
	}
	// Approximate distance: 1 if edit1 of low, else 2 if within edit2, else 3
	low := strings.ToLower(word)
	edit1set := map[string]struct{}{}
	for _, c := range edit1Candidates(low) {
		edit1set[c] = struct{}{}
	}
	out := make([]struct {
		Word   string
		Weight int
	}, 0, len(sugs))
	for _, s := range sugs {
		sl := strings.ToLower(s)
		dist := 2
		if _, ok := edit1set[sl]; ok {
			dist = 1
		} else if maxEdit >= 3 {
			// cheap: if any edit1 of s is edit1 of word → distance 2
			dist = 3
			for _, e := range edit1Candidates(sl) {
				if _, ok := edit1set[e]; ok {
					dist = 2
					break
				}
			}
		}
		freq := d.GetFrequency(sl)
		if freq < 0 {
			freq = 0
		}
		w := dist*FreqRanges + FreqRanges - freq - 1
		out = append(out, struct {
			Word   string
			Weight int
		}{Word: s, Weight: w})
	}
	// sort by weight ascending
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j].Weight < out[i].Weight {
				out[i], out[j] = out[j], out[i]
			}
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
