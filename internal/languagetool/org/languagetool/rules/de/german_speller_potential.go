package de

import (
	"regexp"
	"strings"
)

// Ports GermanSpellerRule.ignorePotentiallyMisspelledWord: early gates + compound
// tokenize + processTwoPart/ThreePart. POS-heavy arms need TagPOS/LemmaOf (wired
// from added.txt when present; full german.dict still optional).

// German compound-ignore word length bounds (Java MIN_WORD_LENGTH / MAX_WORD_LENGTH).
const (
	germanPotentialMinWordLength = 5
	germanPotentialMaxWordLength = 40
)

var (
	// CAMEL_CASE: any lower-upper or multi-upper-then-lower transition.
	reCamelCase = regexp.MustCompile(`.*(\p{Ll}\p{Lu}|\p{Lu}{2,}\p{Ll}).*`)
	// COMPOUND_TYPOS / COMPOUND_END_TYPOS — likely typos, never auto-accept as compounds.
	reCompoundTypos    = regexp.MustCompile(`^(?:[Ee]mail|[Ii]reland|[Kk]reissaal|[Mm]akeup|[Ss]tandart).*`)
	reCompoundEndTypos = regexp.MustCompile(`(?:gruße|schaf(?:s|en)?)$`)
	// GENDER_STAR pieces (Java lookbehind (?<=\w)In not in RE2 — use capture).
	reGenderStarBinnenI = regexp.MustCompile(`([0-9A-Za-zÄÖÜäöüß])In`)
	reGenderStarMarker  = regexp.MustCompile(`[\*:_]in|/-in`)
)

// isValidWordLengthForPotential ports isValidWordLength: true means OUTSIDE the
// compound-ignore window (too short or too long) → do not ignore via potential path.
func isValidWordLengthForPotential(word string) bool {
	n := len([]rune(word))
	return n <= germanPotentialMinWordLength || n >= germanPotentialMaxWordLength
}

// isProbablyTypo ports GermanSpellerRule.isProbablyTypo.
func isProbablyTypo(word string) bool {
	return reCompoundTypos.MatchString(word) || reCompoundEndTypos.MatchString(word)
}

// isValidCamelCase ports GermanSpellerRule.isValidCamelCase (true if NOT camelCase).
func isValidCamelCase(input string) bool {
	return !reCamelCase.MatchString(input)
}

// genderStarNormalize ports GENDER_STAR.replaceFirst(..., "in") once.
// Alternation order: (?<=\w)In → keep letter + "in"; else [\*:_]in|/-in → "in".
func genderStarNormalize(word string) string {
	if loc := reGenderStarBinnenI.FindStringSubmatchIndex(word); loc != nil {
		// full [0:1], group1 letter [2:3]
		letter := word[loc[2]:loc[3]]
		return word[:loc[0]] + letter + "in" + word[loc[1]:]
	}
	if loc := reGenderStarMarker.FindStringIndex(word); loc != nil {
		return word[:loc[0]] + "in" + word[loc[1]:]
	}
	return word
}

// removeTrailingSAndHyphen ports removeTrailingS + removeTrailingHyphen.
func removeTrailingSAndHyphen(part string) string {
	part = strings.TrimSuffix(part, "-")
	part = strings.TrimSuffix(part, "s")
	return part
}

// IgnorePotentiallyMisspelledWord ports GermanSpellerRule.ignorePotentiallyMisspelledWord:
// early gates, then CompoundTokenize + ProcessTwoPart/ThreePart when wired.
// Without CompoundTokenize/TagPOS resources, compound accept stays fail-closed (false).
func (r *GermanSpellerRule) IgnorePotentiallyMisspelledWord(word string) bool {
	if r == nil || word == "" {
		return false
	}
	wordNoDot := cutOffDot(word)
	// Java: if (isValidWordLength(word) || startsWithLowercase || isProhibited...) return false
	if isValidWordLengthForPotential(word) || startsWithLowercase(word) ||
		r.IsProhibited(word) || r.IsProhibited(wordNoDot) {
		return false
	}
	if isProbablyTypo(word) {
		return false
	}
	wordNoDotOrg := wordNoDot
	wordNoDotNorm := genderStarNormalize(wordNoDot)
	if !isValidCamelCase(wordNoDotNorm) {
		return false
	}

	// CompoundTokenizer path (Java compoundTokenizer / nonStrict fallback).
	if r.CompoundTokenize == nil && r.CompoundTokenizeNonStrict == nil {
		return false
	}
	var parts []string
	if r.CompoundTokenize != nil {
		parts = r.CompoundTokenize(wordNoDotNorm)
	}
	if len(parts) <= 1 && r.CompoundTokenizeNonStrict != nil {
		parts = r.CompoundTokenizeNonStrict(wordNoDotNorm)
	}
	// If at least one element equals "s", append to predecessor
	parts = avoidInfixSAsSingleToken(parts)

	// Hyphenated compounds: misspelled segments, split hyphen tokens, restore hyphens
	if strings.Contains(wordNoDotNorm, "-") {
		splitByHyphen := strings.Split(wordNoDotNorm, "-")
		if len(splitByHyphen) > 0 {
			lastPart := splitByHyphen[len(splitByHyphen)-1]
			// e.g. "Implementierungs-pflicht" — last part lower but is noun when uppercased
			if !r.isNoun(lastPart) && r.isNoun(uppercaseFirstChar(lastPart)) {
				return false
			}
		}
		for _, w := range splitByHyphen {
			if w == "" {
				continue
			}
			if r.IsMisspelled(w) && r.IsMisspelled(removeTrailingSAndHyphen(w)) {
				return false
			}
		}
		// "Wacht" + "ums-pistole" → "Wacht" + "ums" + "pistole"
		parts = splitPartsByHyphen(parts)
		// Hyphens often removed by tokenizer — restore for processTwoPart
		parts = restoreRemovedHyphens(parts, wordNoDotNorm)
	}

	if len(parts) <= 1 {
		return false
	}
	if !isValidPartLength(parts) {
		return false
	}
	if r.isOldSpelling(parts) {
		return false
	}
	// Gender-neutral check against original surface (before gender normalize)
	if hasGender2Star2(wordNoDotOrg) {
		if !r.isValidGenderNeutralWord(parts, wordNoDotOrg) {
			return false
		}
	}
	switch len(parts) {
	case 2:
		return r.ProcessTwoPartCompounds(parts[0], parts[1])
	case 3:
		return r.ProcessThreePartCompound(parts)
	default:
		return false
	}
}

// splitPartsByHyphen ports GermanSpellerRule.splitPartsByHyphen:
// any part containing "-" is expanded into separate tokens.
func splitPartsByHyphen(parts []string) []string {
	if len(parts) == 0 {
		return parts
	}
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if !strings.Contains(p, "-") {
			out = append(out, p)
			continue
		}
		for _, sub := range strings.Split(p, "-") {
			if sub != "" {
				out = append(out, sub)
			}
		}
	}
	return out
}

// restoreRemovedHyphens ports GermanSpellerRule.restoreRemovedHyphens:
// append "-" to a part when the original word has a hyphen immediately after
// that part's span (Java String indices ≈ runes for DE BMP).
func restoreRemovedHyphens(parts []string, word string) []string {
	if len(parts) == 0 || word == "" {
		return parts
	}
	wrunes := []rune(word)
	var hyphenPositions []int
	for i, r := range wrunes {
		if r == '-' {
			hyphenPositions = append(hyphenPositions, i)
		}
	}
	if len(hyphenPositions) == 0 {
		return parts
	}
	out := make([]string, 0, len(parts))
	currentPos := 0
	for _, token := range parts {
		tok := token
		tlen := len([]rune(token))
		for _, hp := range hyphenPositions {
			if hp >= currentPos && hp == currentPos+tlen {
				tok = token + "-"
				break
			}
		}
		out = append(out, tok)
		// Java advances by token.length() after optional "-" append
		currentPos += len([]rune(tok))
	}
	return out
}

// FilterProhibitedSuggestions ports SpellingCheckRule.filterSuggestions core:
// drop replacements that isProhibited.
func (r *GermanSpellerRule) FilterProhibitedSuggestions(sugs []string) []string {
	if r == nil || len(sugs) == 0 {
		return sugs
	}
	var out []string
	for _, s := range sugs {
		if r.IsProhibited(s) {
			continue
		}
		if !r.AcceptSuggestion(s) {
			continue
		}
		out = append(out, s)
	}
	return out
}

