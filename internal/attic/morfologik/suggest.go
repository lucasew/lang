package morfologik

import (
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

// SuggestEdits ports Speller.findReplacements (maxEdit=1) word list.
// max caps the result size (0 → 8). Case fold is MorfologikSpeller layer, not Speller.
func (d *Dictionary) SuggestEdits(word string, max int) []string {
	return d.SuggestEditsMax(word, max, 1)
}

// SuggestEditsMax ports Speller.findReplacements with maxEditDistance.
func (d *Dictionary) SuggestEditsMax(word string, maxResults, maxEdit int) []string {
	return d.SuggestEditsMaxOpts(word, maxResults, maxEdit, SuggestOpts{})
}

// SuggestEditsMaxOpts returns findReplacementCandidates words.
// opt is ignored for production path — Java Speller uses DictionaryMetadata only
// (loaded from .info). Kept for API compatibility with older call sites.
func (d *Dictionary) SuggestEditsMaxOpts(word string, maxResults, maxEdit int, opt SuggestOpts) []string {
	_ = opt
	if d == nil || word == "" {
		return nil
	}
	if maxResults <= 0 {
		maxResults = 8
	}
	if maxEdit < 1 {
		maxEdit = 1
	}
	cds := d.FindReplacementCandidates(word, maxEdit)
	if len(cds) == 0 {
		return nil
	}
	out := make([]string, 0, len(cds))
	for _, c := range cds {
		if c.Word == "" {
			continue
		}
		out = append(out, c.Word)
		if len(out) >= maxResults {
			break
		}
	}
	return out
}

// WeightedEditSuggestions returns CandidateData distances (Java WeightedSuggestion weights).
func (d *Dictionary) WeightedEditSuggestions(word string, maxResults, maxEdit int) []struct {
	Word   string
	Weight int
} {
	return d.WeightedEditSuggestionsOpts(word, maxResults, maxEdit, SuggestOpts{})
}

// WeightedEditSuggestionsOpts ports Speller.findReplacementCandidates → weighted list.
// opt ignored (DictionaryMetadata from .info is king).
func (d *Dictionary) WeightedEditSuggestionsOpts(word string, maxResults, maxEdit int, opt SuggestOpts) []struct {
	Word   string
	Weight int
} {
	_ = opt
	if d == nil || word == "" {
		return nil
	}
	if maxResults <= 0 {
		maxResults = 8
	}
	if maxEdit < 1 {
		maxEdit = 1
	}
	cds := d.FindReplacementCandidates(word, maxEdit)
	if len(cds) == 0 {
		return nil
	}
	if maxResults > 0 && len(cds) > maxResults {
		cds = cds[:maxResults]
	}
	out := make([]struct {
		Word   string
		Weight int
	}, 0, len(cds))
	for _, c := range cds {
		out = append(out, struct {
			Word   string
			Weight int
		}{Word: c.Word, Weight: c.Distance})
	}
	// already sorted by FindReplacementCandidates; ensure stable non-decreasing
	for i := 0; i < len(out); i++ {
		for j := i + 1; j < len(out); j++ {
			if out[j].Weight < out[i].Weight {
				out[i], out[j] = out[j], out[i]
			}
		}
	}
	return out
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

// stripDiacritic returns the first non-mark code point (test helper / NFD base).
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
