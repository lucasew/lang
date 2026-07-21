package de

import (
	"regexp"
	"strings"
)

// Patterns from GermanSpellerRule for candidate / language filters.
var (
	// HYPHENED_UPPER_WORD: [A-ZÖÄÜ][a-zöäüß]+-[\\-\\s]?[a-zöäüß]+
	reHyphenedUpperWord = regexp.MustCompile(`^[A-ZÖÄÜ][a-zöäüß]+-[\-\s]?[a-zöäüß]+$`)
	// HYPHENED_WORD: [a-zöäüß]+-[\\-\\s][A-ZÖÄÜa-zöäüß]+
	reHyphenedWord = regexp.MustCompile(`^[a-zöäüß]+-[\-\s][A-ZÖÄÜa-zöäüß]+$`)
	// WORD_WITH_PUNCT: \w\p{Punct}?  (single letter + optional punct token)
	reWordWithPunct = regexp.MustCompile(`^[\p{L}\p{N}_]\p{P}?$`)
	// STARTING_WITH_SINGLE_CHAR: \p{L} \p{L}+
	reStartingWithSingleChar = regexp.MustCompile(`^\p{L} \p{L}+$`)
)

// GetCandidates ports GermanSpellerRule.getCandidates.
// Prefer CompoundTokenizeAll (dictionary getAllSplits stand-in); else single
// CompoundTokenize / non-strict fallback. Does not invent splits outside lexicon.
func (r *GermanSpellerRule) GetCandidates(word string) []string {
	if r == nil || word == "" {
		return nil
	}
	var partLists [][]string
	if r.CompoundTokenizeAll != nil {
		partLists = r.CompoundTokenizeAll(word)
	}
	if len(partLists) == 0 && r.CompoundTokenize != nil {
		parts := r.CompoundTokenize(word)
		if len(parts) > 1 {
			partLists = append(partLists, parts)
		}
	}
	if len(partLists) == 0 && r.CompoundTokenizeNonStrict != nil {
		parts := r.CompoundTokenizeNonStrict(word)
		if len(parts) > 1 {
			partLists = append(partLists, parts)
		}
	}
	var candidates []string
	seen := map[string]struct{}{}
	add := func(s string) {
		if s == "" {
			return
		}
		if _, ok := seen[s]; ok {
			return
		}
		seen[s] = struct{}{}
		candidates = append(candidates, s)
	}
	for _, parts := range partLists {
		tmp := r.getCandidatesFromParts(parts)
		for _, k := range tmp {
			// avoid e.g. "Direkt-weg", "Geheimnis-s-voll"
			if reHyphenedUpperWord.MatchString(k) || reHyphenedWord.MatchString(k) {
				continue
			}
			if strings.Contains(k, "-s-") {
				continue
			}
			if !strings.HasSuffix(word, "-") && strings.HasSuffix(k, "-") {
				continue
			}
			add(k)
		}
		if len(parts) == 2 {
			// e.g. "inneremedizin" -> "innere Medizin", "gleichgroß" -> "gleich groß"
			add(parts[0] + " " + parts[1])
			if r.isNounOrProperNoun(uppercaseFirstChar(parts[1])) {
				add(parts[0] + " " + uppercaseFirstChar(parts[1]))
			}
			if !strings.HasSuffix(parts[0], "s") {
				// Einzahlungschein -> Einzahlungsschein
				add(parts[0] + "s" + parts[1])
			}
			// Java: parts.get(1).startsWith("s") && parts.get(1).length() > 1 → substring(1)
			if strings.HasPrefix(parts[1], "s") && utf16LenDE(parts[1]) > 1 {
				// Ordnungshütter -> Ordnungshüter (split as Ordnung + shütter)
				rest := substringByUTF16(parts[1], 1, utf16LenDE(parts[1]))
				for _, c := range r.getCandidatesFromParts([]string{parts[0] + "s", rest}) {
					add(c)
				}
			}
		}
	}
	return candidates
}

// getCandidatesFromParts ports CompoundAwareHunspellRule.getCandidates(List parts).
// Uses FilterDictSuggest as the Morfologik multi-speller stand-in.
func (r *GermanSpellerRule) getCandidatesFromParts(parts []string) []string {
	if r == nil || len(parts) == 0 || !FilterDictAvailable() {
		return nil
	}
	var candidates []string
	for partCount, part := range parts {
		if !r.IsMisspelled(part) {
			continue
		}
		doUpperCase := partCount > 0 && !startsWithUppercase(part)
		probe := part
		if doUpperCase {
			probe = uppercaseFirstChar(part)
		}
		suggestions := FilterDictSuggest(probe)
		if len(suggestions) == 0 {
			if doUpperCase {
				suggestions = FilterDictSuggest(lowercaseFirstChar(part))
			} else {
				suggestions = FilterDictSuggest(part)
			}
		}
		appendS := false
		if doUpperCase && strings.HasSuffix(part, "s") {
			// maybe infix-s
			base := strings.TrimSuffix(part, "s")
			suggestions = append(suggestions, FilterDictSuggest(base)...)
			appendS = true
		}
		for _, suggestion := range suggestions {
			// Java: if (appendS) { suggestion += "s"; } — applied to every suggestion
			// when part is uppercased and ends with "s" (infix-s probe path).
			sug := suggestion
			if appendS {
				sug = suggestion + "s"
			}
			partsCopy := append([]string(nil), parts...)
			// Java: parts.get(partCount).startsWith("-") && length() > 1
			// → "-" + uppercaseFirstChar(suggestion.substring(1))
			if partCount > 0 && strings.HasPrefix(parts[partCount], "-") && utf16LenDE(parts[partCount]) > 1 {
				rest := ""
				if utf16LenDE(sug) > 1 {
					rest = substringByUTF16(sug, 1, utf16LenDE(sug))
				} else if utf16LenDE(sug) == 1 {
					rest = ""
				}
				partsCopy[partCount] = "-" + uppercaseFirstChar(rest)
			} else if partCount > 0 && !strings.HasSuffix(parts[partCount-1], "-") {
				partsCopy[partCount] = strings.ToLower(sug)
			} else {
				partsCopy[partCount] = sug
			}
			candidate := strings.Join(partsCopy, "")
			if !r.IsMisspelled(candidate) {
				candidates = append(candidates, candidate)
			}
			// Arbeidszimmer -> Arbeitszimmer:
			// Java: suggestion.substring(0, suggestion.length()-1) when ends with "-"
			if partCount < len(parts)-1 && strings.HasSuffix(part, "s") && strings.HasSuffix(sug, "-") {
				if utf16LenDE(sug) >= 1 {
					partsCopy[partCount] = substringByUTF16(sug, 0, utf16LenDE(sug)-1)
				}
				infixCandidate := strings.Join(partsCopy, "")
				if !r.IsMisspelled(infixCandidate) {
					candidates = append(candidates, infixCandidate)
				}
			}
		}
	}
	return candidates
}

// getCorrectWords ports CompoundAwareHunspellRule.getCorrectWords:
// keep phrases where every whitespace token is spelled correctly.
func (r *GermanSpellerRule) getCorrectWords(wordsOrPhrases []string) []string {
	if r == nil {
		return nil
	}
	var result []string
	for _, wordOrPhrase := range wordsOrPhrases {
		words := strings.Fields(wordOrPhrase)
		if len(words) == 0 {
			words = []string{wordOrPhrase}
		}
		ok := true
		for _, w := range words {
			if r.IsMisspelled(w) {
				ok = false
				break
			}
		}
		if ok {
			result = append(result, wordOrPhrase)
		}
	}
	return result
}

// GetFilteredSuggestions ports GermanSpellerRule.getFilteredSuggestions.
// Fail-closed POS: without TagPOS, phrase filters that need POS are skipped
// (phrase kept — same as Java default when tagger would accept; we only drop
// when TagPOS confirms the bad patterns).
func (r *GermanSpellerRule) GetFilteredSuggestions(wordsOrPhrases []string) []string {
	if r == nil {
		return wordsOrPhrases
	}
	var result []string
	for _, wordOrPhrase := range wordsOrPhrases {
		words := strings.Fields(wordOrPhrase)
		if len(words) == 0 {
			words = []string{wordOrPhrase}
		}
		drop := false
		if r.TagPOS != nil && len(words) >= 2 &&
			r.isAdjOrNounOrUnknown(words[0]) && r.isNounOrUnknown(words[1]) &&
			startsWithUppercase(words[0]) && startsWithUppercase(words[1]) {
			// ignore "Release Prozess"
			drop = true
		} else if r.TagPOS != nil && len(words) == 2 &&
			r.isAdjBaseForm(words[0]) && !startsWithUppercase(words[0]) && r.isSubVerInf(words[1]) {
			// filter "groß Denken"
			drop = true
		}
		if !drop {
			result = append(result, wordOrPhrase)
		}
	}
	return result
}

func (r *GermanSpellerRule) isNounOrUnknown(word string) bool {
	if r == nil || r.TagPOS == nil {
		return false
	}
	tags := r.TagPOS(word)
	if len(tags) == 0 {
		return true // isPosTagUnknown
	}
	for _, t := range tags {
		if strings.HasPrefix(t, "SUB") {
			return true
		}
	}
	return false
}

func (r *GermanSpellerRule) isAdjOrNounOrUnknown(word string) bool {
	if r == nil || r.TagPOS == nil {
		return false
	}
	tags := r.TagPOS(word)
	if len(tags) == 0 {
		return true
	}
	for _, t := range tags {
		if strings.HasPrefix(t, "SUB") || strings.HasPrefix(t, "ADJ") {
			return true
		}
	}
	return false
}

func (r *GermanSpellerRule) isNounOrProperNoun(word string) bool {
	if r == nil || r.TagPOS == nil {
		return false
	}
	for _, t := range r.TagPOS(word) {
		if strings.HasPrefix(t, "SUB") || strings.HasPrefix(t, "EIG") {
			return true
		}
	}
	return false
}

func (r *GermanSpellerRule) isSubVerInf(word string) bool {
	if r == nil || r.TagPOS == nil {
		return false
	}
	for _, t := range r.TagPOS(word) {
		// Java: matchesPosTagRegex("SUB:.*:INF")
		if strings.HasPrefix(t, "SUB:") && strings.Contains(t, ":INF") {
			return true
		}
	}
	return false
}

func (r *GermanSpellerRule) isAdjBaseForm(word string) bool {
	if r == nil || r.TagPOS == nil {
		return false
	}
	for _, t := range r.TagPOS(word) {
		if strings.HasPrefix(t, "ADJ:PRD:GRU") {
			return true
		}
	}
	return false
}

// FilterForLanguage ports GermanSpellerRule.filterForLanguage.
func (r *GermanSpellerRule) FilterForLanguage(suggestions []string) []string {
	if r == nil || len(suggestions) == 0 {
		return suggestions
	}
	var out []string
	for _, s := range suggestions {
		if r.LanguageVariant == "CH" {
			s = strings.ReplaceAll(s, "ß", "ss")
		}
		// Remove suggestions like "Mafiosi s" and "Mafiosi s."
		drop := false
		for _, k := range strings.Fields(s) {
			if reWordWithPunct.MatchString(k) {
				drop = true
				break
			}
		}
		if drop {
			continue
		}
		// Java: s.length() > 1 && s.startsWith("-") — UTF-16 length
		if utf16LenDE(s) > 1 && strings.HasPrefix(s, "-") {
			continue
		}
		out = append(out, s)
	}
	return out
}

// postFilterGetSuggestions ports the stream filters at the end of
// GermanSpellerRule.getSuggestions (after acceptSuggestion / period fix).
func postFilterGetSuggestions(word string, suggestions []string) []string {
	var out []string
	for _, k := range suggestions {
		if k == word {
			continue
		}
		if strings.HasSuffix(k, "-") && !strings.HasSuffix(word, "-") {
			continue
		}
		if reStartingWithSingleChar.MatchString(k) {
			continue
		}
		out = append(out, k)
	}
	return out
}

// interleaveSuggestions ports CompoundAwareHunspellRule mixing of suggestion lists.
func interleaveSuggestions(lists ...[]string) []string {
	max := 0
	for _, l := range lists {
		if len(l) > max {
			max = len(l)
		}
	}
	var out []string
	for i := 0; i < max; i++ {
		for _, l := range lists {
			if i < len(l) {
				out = append(out, l[i])
			}
		}
	}
	return out
}

func dedupeSuggestions(in []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, s := range in {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

const maxGermanSuggestions = 20
