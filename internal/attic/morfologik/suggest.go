package morfologik

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// FREQ_RANGES ports morfologik Speller.FREQ_RANGES ('Z'-'A'+1 = 26).
const FreqRanges = 26

// SuggestOpts ports Speller.areEqual / edit-search options from DictionaryMetadata.
type SuggestOpts struct {
	// IgnoreDiacritics ports fsa.dict.speller.ignore-diacritics (EN true).
	IgnoreDiacritics bool
	// ConvertCase ports fsa.dict.speller.convert-case (used inside areEqual diacritic fold).
	ConvertCase bool
	// EquivalentChars ports fsa.dict.speller.equivalent-chars (from → list of to).
	// Speller.areEqual only checks map[from].contains(to), not reverse.
	EquivalentChars map[rune][]rune
	// SymmetricEquivalent enables reverse MAP lookup for invent edit-candidate generation only
	// (not Java areEqual). Leave false for SpellerED / findRepl.
	SymmetricEquivalent bool
}

// SuggestEdits returns dictionary words within one Damerau-Levenshtein edit of word
// by generating candidates and testing Contains (works for large CFSA2 dicts).
// max caps the result size (0 → 8).
func (d *Dictionary) SuggestEdits(word string, max int) []string {
	return d.SuggestEditsMax(word, max, 1)
}

// SuggestEditsMax ports Speller suggestions with maxEditDistance (1, 2, or 3).
func (d *Dictionary) SuggestEditsMax(word string, maxResults, maxEdit int) []string {
	return d.SuggestEditsMaxOpts(word, maxResults, maxEdit, SuggestOpts{})
}

// SuggestEditsMaxOpts is SuggestEditsMax with areEqual-related options.
func (d *Dictionary) SuggestEditsMaxOpts(word string, maxResults, maxEdit int, opt SuggestOpts) []string {
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
	cands1 := edit1CandidatesOpts(low, opt)
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
		for _, c2 := range edit1CandidatesOpts(c, opt) {
			if addIfKnown(c2) && len(out) >= maxResults {
				return out
			}
		}
	}
	if maxEdit < 3 || len(out) >= maxResults {
		return out
	}

	// distance 3: one more expansion round on a smaller frontier
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
		for _, c2 := range edit1CandidatesOpts(c, opt) {
			if d.Contains(c2) {
				continue
			}
			inner++
			if inner > 30 {
				break
			}
			for _, c3 := range edit1CandidatesOpts(c2, opt) {
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
	return d.WeightedEditSuggestionsOpts(word, maxResults, maxEdit, SuggestOpts{})
}

// WeightedEditSuggestionsOpts is WeightedEditSuggestions with Speller.areEqual options.
func (d *Dictionary) WeightedEditSuggestionsOpts(word string, maxResults, maxEdit int, opt SuggestOpts) []struct {
	Word   string
	Weight int
} {
	sugs := d.SuggestEditsMaxOpts(word, maxResults, maxEdit, opt)
	if len(sugs) == 0 {
		return nil
	}
	// Approximate distance: 1 if edit1 of low, else 2 if within edit2, else 3
	// Free diacritic/equivalent diffs count as distance 0 (Java areEqual).
	low := strings.ToLower(word)
	edit1set := map[string]struct{}{}
	for _, c := range edit1CandidatesOpts(low, opt) {
		edit1set[c] = struct{}{}
	}
	out := make([]struct {
		Word   string
		Weight int
	}, 0, len(sugs))
	for _, s := range sugs {
		sl := strings.ToLower(s)
		dist := 2
		if freeEqualUnderOpts(low, sl, opt) {
			dist = 0
		} else if _, ok := edit1set[sl]; ok {
			dist = 1
		} else if maxEdit >= 3 {
			dist = 3
			for _, e := range edit1CandidatesOpts(sl, opt) {
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

// freeEqualUnderOpts reports whether a and b match under areEqual (same length, free chars).
func freeEqualUnderOpts(a, b string, opt SuggestOpts) bool {
	ar, br := []rune(a), []rune(b)
	if len(ar) != len(br) {
		return false
	}
	for i := range ar {
		if !runesEqualUnderOpts(ar[i], br[i], opt) {
			return false
		}
	}
	return true
}

// runesEqualUnderOpts ports Speller.areEqual for a single character pair (Java 2.2.0).
func runesEqualUnderOpts(x, y rune, opt SuggestOpts) bool {
	if x == y {
		return true
	}
	if opt.EquivalentChars != nil {
		if list, ok := opt.EquivalentChars[x]; ok {
			for _, c := range list {
				if c == y {
					return true
				}
			}
		}
		// invent edit-gen only (not Speller.areEqual)
		if opt.SymmetricEquivalent {
			if list, ok := opt.EquivalentChars[y]; ok {
				for _, c := range list {
					if c == x {
						return true
					}
				}
			}
		}
	}
	if opt.IgnoreDiacritics {
		xn := nfdFirst(x)
		yn := nfdFirst(y)
		if xn == yn {
			return true
		}
		if opt.ConvertCase && unicode.IsLetter(xn) {
			if unicode.IsLower(xn) != unicode.IsLower(yn) {
				return unicode.ToLower(xn) == unicode.ToLower(yn)
			}
		}
		return xn == yn
	}
	return false
}

// nfdFirst ports Normalizer.normalize(Character.toString(x), NFD).charAt(0).
func nfdFirst(r rune) rune {
	s := norm.NFD.String(string(r))
	if s == "" {
		return r
	}
	// UTF-16 charAt(0) for BMP = first rune of NFD string
	for _, c := range s {
		return c
	}
	return r
}

// stripDiacritic returns the first non-mark code point (legacy edit-gen helper).
func stripDiacritic(r rune) rune {
	s := norm.NFD.String(string(r))
	for _, c := range s {
		if unicode.Is(unicode.Mn, c) {
			continue
		}
		return c
	}
	return r
}

// latinDiacriticLetters are common accented Latin letters tried when IgnoreDiacritics
// (so candidate-gen can reach café-style dictionary forms without FSA walk).
const latinDiacriticLetters = "àáâãäåāăąèéêëēĕėęěìíîïĩīĭįòóôõöøōŏőùúûüũūŭůűýÿćĉċčďđĝğġģĥħĵķĺļľŀłńņňŋśŝşšţťŧŵŷźżžçñæœß"

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

// edit1Candidates generates lowercase distance-1 candidates (ASCII letters + apostrophe).
func edit1Candidates(word string) []string {
	return edit1CandidatesOpts(word, SuggestOpts{})
}

// edit1CandidatesOpts expands the replace/insert alphabet for ignore-diacritics and
// equivalent-chars (Java Speller.areEqual surface for candidate generation).
func edit1CandidatesOpts(word string, opt SuggestOpts) []string {
	letters := []rune("abcdefghijklmnopqrstuvwxyz'")
	if opt.IgnoreDiacritics {
		letters = append(letters, []rune(latinDiacriticLetters)...)
	}
	if opt.EquivalentChars != nil {
		seen := map[rune]struct{}{}
		for _, r := range letters {
			seen[r] = struct{}{}
		}
		for from, tos := range opt.EquivalentChars {
			if _, ok := seen[from]; !ok {
				letters = append(letters, from)
				seen[from] = struct{}{}
			}
			for _, t := range tos {
				if _, ok := seen[t]; !ok {
					letters = append(letters, t)
					seen[t] = struct{}{}
				}
			}
		}
	}
	w := []rune(word)
	n := len(w)
	out := make([]string, 0, n*len(letters)*2+n)
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
		// equivalent-char free replacements for the char at i
		if opt.EquivalentChars != nil {
			if list, ok := opt.EquivalentChars[w[i]]; ok {
				for _, c := range list {
					if c == w[i] {
						continue
					}
					rw := append([]rune{}, w...)
					rw[i] = c
					out = append(out, string(rw))
				}
			}
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
