package de

import "strings"

// SortSuggestionByQuality ports GermanSpellerRule.sortSuggestionByQuality.
// Lemma/POS filtering requires TagPOS+LemmaOf; without them that stage is skipped
// (keep all forms). Language-model split ranking is omitted when no LM (Java
// else branch: treat space suggestions as top).
func (r *GermanSpellerRule) SortSuggestionByQuality(misspelling string, suggestions []string) []string {
	if r == nil || len(suggestions) == 0 {
		return suggestions
	}

	filtered := suggestions
	// Java: misspelling.length() > 1 (UTF-16) before lemma/POS filter stage
	if r.TagPOS != nil && utf16LenDE(misspelling) > 1 {
		filtered = r.filterSameLemmaInflections(misspelling, suggestions)
	}

	// Boost: case-only differ, and suggestions containing space
	var result []string
	var top []string
	for _, suggestion := range filtered {
		if strings.EqualFold(misspelling, suggestion) {
			top = append(top, suggestion)
			continue
		}
		if strings.Contains(suggestion, " ") {
			// Java: languageModel may demote; without LM always top
			top = append(top, suggestion)
			continue
		}
		result = append(result, suggestion)
	}
	return append(top, result...)
}

// filterSameLemmaInflections ports the ADJ/SUB/PA same-lemma filter at the
// start of sortSuggestionByQuality.
func (r *GermanSpellerRule) filterSameLemmaInflections(misspelling string, suggestions []string) []string {
	// Java: misspelling.substring(misspelling.length() - 2) — UTF-16 units
	if utf16LenDE(misspelling) < 2 {
		return suggestions
	}
	suffix2 := substringByUTF16(misspelling, utf16LenDE(misspelling)-2, utf16LenDE(misspelling))

	formToAccept := ""
	lemmaToFilter := ""
	for _, sug := range suggestions {
		tags := r.TagPOS(sug)
		if len(tags) == 0 {
			continue
		}
		okPOS := false
		for _, t := range tags {
			if strings.HasPrefix(t, "ADJ") || strings.HasPrefix(t, "SUB") ||
				strings.HasPrefix(t, "PA1:") || strings.HasPrefix(t, "PA2:") {
				okPOS = true
				break
			}
		}
		if !okPOS {
			continue
		}
		// Java: readings.getToken().endsWith(suffix2)
		if !strings.HasSuffix(sug, suffix2) {
			continue
		}
		formToAccept = sug
		if r.LemmaOf != nil {
			lemmaToFilter = r.LemmaOf(sug)
		}
		break
	}
	if formToAccept == "" || lemmaToFilter == "" {
		return suggestions
	}

	var filtered []string
	seen := map[string]struct{}{}
	for _, sug := range suggestions {
		if sug == formToAccept {
			if _, ok := seen[sug]; !ok {
				filtered = append(filtered, sug)
				seen[sug] = struct{}{}
			}
			continue
		}
		lem := ""
		if r.LemmaOf != nil {
			lem = r.LemmaOf(sug)
		}
		// keep if not same lemma
		if lem == "" || lem != lemmaToFilter {
			if _, ok := seen[sug]; !ok {
				filtered = append(filtered, sug)
				seen[sug] = struct{}{}
			}
		}
	}
	if len(filtered) == 0 {
		return suggestions
	}
	return filtered
}
