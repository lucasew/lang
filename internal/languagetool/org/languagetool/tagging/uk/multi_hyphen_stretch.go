package uk

import (
	"regexp"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// abbrPosRE ports CompoundTagger.ABBR_PATTERN .*abbr.*
var abbrPosRE = regexp.MustCompile(`(?s).*abbr.*`)

// collapseStretch ports CompoundTagger.collapseStretch.
// Java STRETCH_PATTERN uses backrefs (unsupported in RE2) — equivalent manual pass twice,
// then strip remaining hyphens; re-capitalize if original was capitalized.
func collapseStretch(word string) string {
	capitalized := tools.IsCapitalizedWord(word)
	merged := collapseStretchOnce(strings.ToLower(word))
	merged = collapseStretchOnce(merged)
	merged = strings.ReplaceAll(merged, "-", "")
	if capitalized {
		merged = capitalizeFirst(merged)
	}
	return merged
}

// collapseStretchOnce replaces X+ - X+ (same letter, case-insensitive) with single X.
// Ports Java: ([а-іяїєґА-ЯІЇЄҐ])\1*-\1+ → $1
func collapseStretchOnce(s string) string {
	rs := []rune(s)
	if len(rs) < 3 {
		return s
	}
	var out []rune
	i := 0
	for i < len(rs) {
		// try match starting at i: letter, optional same letters, hyphen, one+ same letters
		r0 := rs[i]
		if !isUkrLetter(r0) {
			out = append(out, r0)
			i++
			continue
		}
		base := unicodeToLower(r0)
		j := i + 1
		for j < len(rs) && unicodeToLower(rs[j]) == base {
			j++
		}
		if j >= len(rs) || rs[j] != '-' {
			out = append(out, rs[i])
			i++
			continue
		}
		// need at least one same letter after hyphen
		k := j + 1
		if k >= len(rs) || unicodeToLower(rs[k]) != base {
			out = append(out, rs[i])
			i++
			continue
		}
		for k < len(rs) && unicodeToLower(rs[k]) == base {
			k++
		}
		// collapse run to single base letter (lower, capital later)
		out = append(out, base)
		i = k
	}
	return string(out)
}

func isUkrLetter(r rune) bool {
	switch {
	case r >= 'а' && r <= 'я', r >= 'А' && r <= 'Я':
		return true
	case r == 'і' || r == 'ї' || r == 'є' || r == 'ґ' || r == 'І' || r == 'Ї' || r == 'Є' || r == 'Ґ':
		return true
	}
	return false
}

func unicodeToLower(r rune) rune {
	// Unicode lower for Ukrainian
	if r >= 'А' && r <= 'Я' {
		return r + ('а' - 'А')
	}
	switch r {
	case 'І':
		return 'і'
	case 'Ї':
		return 'ї'
	case 'Є':
		return 'є'
	case 'Ґ':
		return 'ґ'
	}
	return r
}

// tagBothCases ports CompoundTagger.tagBothCases(word, posTagMatcher).
// Tags as-is + lower (or upper if already lower); optional full-match POS filter.
func tagBothCases(word string, tagWord func(string) []tagging.TaggedWord, posRE *regexp.Regexp) []tagging.TaggedWord {
	if tagWord == nil || word == "" {
		return nil
	}
	var out []tagging.TaggedWord
	seen := map[string]struct{}{}
	add := func(w string) {
		for _, tw := range tagWord(w) {
			if posRE != nil && (tw.PosTag == "" || !posMatchesFull(posRE, tw.PosTag)) {
				continue
			}
			key := tw.Lemma + "|" + tw.PosTag
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			out = append(out, tw)
		}
	}
	add(word)
	low := strings.ToLower(word)
	if low != word {
		add(low)
	} else {
		up := capitalizeFirst(word)
		if up != word {
			add(up)
		}
	}
	return out
}

// DynamicMultiHyphenStretchReadings ports doGuessMultiHyphens merge/stretch arms:
//  1. parts.length==3 → EntityReadings (official entities)
//  2. parts>=3, unique>1, first not dash prefix/invalid:
//     merge hyphens away → dict + :alt
//     collapseStretch → dict + :alt
// Intj redup is handled by DynamicIntjRedupReadings.
func DynamicMultiHyphenStretchReadings(token string, tagWord func(string) []tagging.TaggedWord) []*languagetool.AnalyzedToken {
	if token == "" || !strings.Contains(token, "-") {
		return nil
	}
	t := strings.ReplaceAll(token, "–", "-")
	t = strings.ReplaceAll(t, "—", "-")
	t = strings.ReplaceAll(t, "\u2011", "-")
	// Java uses lowerWord.split for set logic, original word for surface/entities
	lowerWord := strings.ToLower(t)
	parts := strings.Split(lowerWord, "-")
	if len(parts) < 2 {
		return nil
	}
	for _, p := range parts {
		if p == "" {
			return nil
		}
	}

	// parts.length == 3 → generateEntities(word)
	if len(parts) == 3 {
		if ents := EntityReadings(token); len(ents) > 0 {
			return ents
		}
		// also try normalized dashes surface
		if token != t {
			if ents := EntityReadings(t); len(ents) > 0 {
				return ents
			}
		}
	}

	// filter out г-г-г (set.size()==1 handled by intj redup)
	uniq := uniqueStrings(parts)
	if len(parts) < 3 || len(uniq) <= 1 {
		return nil
	}
	first := parts[0]
	if IsDashPrefix(first) || IsDashPrefixInvalid(first) {
		return nil
	}

	// ва-ре-ни-ки: strip all hyphens
	merged := strings.ReplaceAll(t, "-", "")
	if tagged := filterAbbrNegative(tagBothCases(merged, tagWord, nil)); len(tagged) > 0 {
		return taggedWordsToSurfaceTokens(token, AddIfNotContainsWords(tagged, ":alt", ""))
	}

	// ду-у-у-же / Та-а-ак
	stretched := collapseStretch(t)
	if stretched != "" && stretched != merged {
		if tagged := filterAbbrNegative(tagBothCases(stretched, tagWord, nil)); len(tagged) > 0 {
			return taggedWordsToSurfaceTokens(token, AddIfNotContainsWords(tagged, ":alt", ""))
		}
	} else if stretched != "" {
		// still try stretch result even if equal after other normalize
		if tagged := filterAbbrNegative(tagBothCases(stretched, tagWord, nil)); len(tagged) > 0 {
			// avoid double if same as merged already returned empty
			return taggedWordsToSurfaceTokens(token, AddIfNotContainsWords(tagged, ":alt", ""))
		}
	}
	return nil
}

func filterAbbrNegative(words []tagging.TaggedWord) []tagging.TaggedWord {
	return Filter2Negative(words, abbrPosRE)
}

func taggedWordsToSurfaceTokens(surface string, words []tagging.TaggedWord) []*languagetool.AnalyzedToken {
	var out []*languagetool.AnalyzedToken
	for _, tw := range words {
		if tw.PosTag == "" {
			continue
		}
		p, l := tw.PosTag, tw.Lemma
		out = append(out, languagetool.NewAnalyzedToken(surface, &p, &l))
	}
	return out
}

