package morfologik

import (
	"strings"
)

// MorfologikSpeller ports org.languagetool.rules.spelling.morfologik.MorfologikSpeller
// as a pluggable dictionary probe + optional suggestion map (binary .dict deferred).
type MorfologikSpeller struct {
	// FileInClassPath is the dictionary resource path (API parity).
	FileInClassPath string
	MaxEditDistance int
	// Words accepted by the speller.
	Words map[string]struct{}
	// Suggestions for misspellings.
	Suggestions map[string][]string
	// ConversionLocale lowercases via strings.ToLower when set.
	ConversionLocale string
}

func NewMorfologikSpeller(fileInClassPath string, maxEditDistance int) *MorfologikSpeller {
	if maxEditDistance < 1 {
		maxEditDistance = 1
	}
	return &MorfologikSpeller{
		FileInClassPath: fileInClassPath,
		MaxEditDistance: maxEditDistance,
		Words:           map[string]struct{}{},
		Suggestions:     map[string][]string{},
	}
}

// AddWord registers an accepted dictionary form.
func (s *MorfologikSpeller) AddWord(word string) {
	if s.Words == nil {
		s.Words = map[string]struct{}{}
	}
	s.Words[word] = struct{}{}
}

// IsMisspelled returns true if word is not in the dictionary.
func (s *MorfologikSpeller) IsMisspelled(word string) bool {
	if s == nil || word == "" {
		return false
	}
	if _, ok := s.Words[word]; ok {
		return false
	}
	// try lowercase form
	low := strings.ToLower(word)
	if low != word {
		if _, ok := s.Words[low]; ok {
			return false
		}
	}
	return true
}

// ConvertsCase reports case-folding acceptance (Java MorfologikSpeller.convertsCase).
// Map speller always converts case via strings.ToLower probe.
func (s *MorfologikSpeller) ConvertsCase() bool { return s != nil }

// GetSuggestions is the Java API alias for FindReplacements.
func (s *MorfologikSpeller) GetSuggestions(word string) []string {
	return s.FindReplacements(word)
}

// FindReplacements returns suggestions for word (map first, then trivial edit-distance peers).
func (s *MorfologikSpeller) FindReplacements(word string) []string {
	if s == nil {
		return nil
	}
	if sug, ok := s.Suggestions[word]; ok {
		return append([]string(nil), sug...)
	}
	// limited: collect dictionary words within MaxEditDistance (small dicts only)
	if len(s.Words) == 0 || len(s.Words) > 5000 {
		return nil
	}
	var out []string
	for w := range s.Words {
		d := editDistance(word, w)
		// exclude exact dictionary form (Java getSuggestions returns empty for known words)
		if d > 0 && d <= s.MaxEditDistance {
			out = append(out, w)
			if len(out) >= 8 {
				break
			}
		}
	}
	return out
}

func editDistance(a, b string) int {
	// simple Levenshtein on runes
	ar, br := []rune(a), []rune(b)
	if len(ar) == 0 {
		return len(br)
	}
	if len(br) == 0 {
		return len(ar)
	}
	prev := make([]int, len(br)+1)
	cur := make([]int, len(br)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(ar); i++ {
		cur[0] = i
		for j := 1; j <= len(br); j++ {
			cost := 1
			if ar[i-1] == br[j-1] {
				cost = 0
			}
			del := prev[j] + 1
			ins := cur[j-1] + 1
			sub := prev[j-1] + cost
			cur[j] = min3(del, ins, sub)
		}
		prev, cur = cur, prev
	}
	return prev[len(br)]
}

func min3(a, b, c int) int {
	if a < b {
		if a < c {
			return a
		}
		return c
	}
	if b < c {
		return b
	}
	return c
}
