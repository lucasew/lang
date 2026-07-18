package en

import (
	"regexp"
	"strings"
	"unicode"

	atticmorfo "github.com/lucasew/lang/internal/attic/morfologik"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// enPunctPCTRE ports EN disambiguation UNKNOWN_PCT: [\.,;:…!\?] → add POS PCT.
var enPunctPCTRE = regexp.MustCompile(`^[\.,;:…!\?]+$`)

// RegisterBinaryEnglishTagger installs lt.TagWord backed by CFSA2 english.dict POS lookup.
// Returns false if the dictionary cannot be opened.
func RegisterBinaryEnglishTagger(lt *languagetool.JLanguageTool, dictPath string) bool {
	if lt == nil || dictPath == "" {
		return false
	}
	d, err := atticmorfo.OpenDictionary(dictPath)
	if err != nil || d == nil {
		return false
	}
	lt.TagWord = BinaryEnglishTagWord(d)
	return true
}

// BinaryEnglishTagWord returns a TagWord inject from an opened POS dictionary.
// Case logic follows Java EnglishTagger.tag: always merge lowercase tags for
// non-lowercase, non-mixed-case tokens (so sentence-start "How" keeps WRB).
func BinaryEnglishTagWord(d *atticmorfo.Dictionary) func(token string) []languagetool.TokenTag {
	if d == nil {
		return nil
	}
	lookup := func(w string) []languagetool.TokenTag {
		forms, err := d.Lookup(w)
		if err != nil || len(forms) == 0 {
			return nil
		}
		out := make([]languagetool.TokenTag, 0, len(forms))
		for _, f := range forms {
			out = append(out, languagetool.TokenTag{POS: f.Tag, Lemma: f.Stem})
		}
		return out
	}
	return func(token string) []languagetool.TokenTag {
		if token == "" {
			return nil
		}
		// Java: typewriter apostrophe so dict entries match.
		word := strings.ReplaceAll(token, "’", "'")
		low := strings.ToLower(word)
		isLower := word == low
		isMixed := englishIsMixedCase(word)
		isAllUpper := word != "" && word == strings.ToUpper(word) && hasLetterEN(word)

		var out []languagetool.TokenTag
		seen := map[string]struct{}{}
		add := func(tags []languagetool.TokenTag) {
			for _, t := range tags {
				key := t.POS + "\x00" + t.Lemma
				if _, ok := seen[key]; ok {
					continue
				}
				seen[key] = struct{}{}
				out = append(out, t)
			}
		}
		// normal case
		add(lookup(word))
		// tag non-lowercase (alluppercase or startuppercase), but not mixed-case,
		// with lowercase word tags (Java EnglishTagger)
		if !isLower && !isMixed {
			add(lookup(low))
		}
		// tag all-uppercase proper nouns (ex. FRANCE) via Title case of lower
		if len(out) == 0 && isAllUpper && low != "" {
			runes := []rune(low)
			title := strings.ToUpper(string(runes[0])) + string(runes[1:])
			if title != word {
				add(lookup(title))
			}
		}
		// walkin' → walking style (Java endsWith "in'")
		if len(out) == 0 && strings.HasSuffix(low, "in'") {
			corrected := word
			if isAllUpper {
				corrected = word[:len(word)-1] + "G"
			} else {
				corrected = word[:len(word)-1] + "g"
			}
			add(lookup(corrected))
			if !isLower && !isMixed {
				add(lookup(strings.ToLower(corrected)))
			}
		}
		// Java disambiguation UNKNOWN_PCT: add PCT on .,;:…!? so grammar
		// patterns postag="…|PCT" match commas (ALL_OF_SUDDEN, etc.).
		if enPunctPCTRE.MatchString(word) {
			add([]languagetool.TokenTag{{POS: "PCT", Lemma: word}})
		}
		return out
	}
}

func hasLetterEN(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) {
			return true
		}
	}
	return false
}

// englishIsMixedCase ports StringTools.isMixedCase: both upper and lower letters,
// not merely initial capital (Title case is not mixed).
func englishIsMixedCase(s string) bool {
	hasUpper, hasLower := false, false
	first := true
	for _, r := range s {
		if !unicode.IsLetter(r) {
			continue
		}
		if unicode.IsUpper(r) {
			if !first {
				// upper after first letter → mixed (iPhone) or ALL_CAPS handled elsewhere
				hasUpper = true
			} else {
				hasUpper = true
			}
		}
		if unicode.IsLower(r) {
			hasLower = true
		}
		first = false
	}
	if !hasUpper || !hasLower {
		return false
	}
	// Title case (first upper, rest lower) is not mixed in LT.
	rs := []rune(s)
	// skip non-letters at start
	i := 0
	for i < len(rs) && !unicode.IsLetter(rs[i]) {
		i++
	}
	if i >= len(rs) {
		return false
	}
	if !unicode.IsUpper(rs[i]) {
		return true // lower then upper somewhere → mixed
	}
	for _, r := range rs[i+1:] {
		if unicode.IsLetter(r) && unicode.IsUpper(r) {
			return true // e.g. iPhone or McDonald - has upper after first
		}
	}
	return false // Title case only
}
