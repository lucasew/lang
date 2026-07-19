package spelling

import "strings"

// FilterSuggestions ports SpellingCheckRule.filterSuggestions for string lists:
//  1. drop isProhibited replacements
//  2. "Name s" → curated "Name" + "Name's" when isProperNoun(Name)
//  3. filterDupes
//  4. filterNoSuggestWords (FilterNoSuggestWordsFn or identity)
//
// TagPOS / IsProperNounFn fail-closed: without them, the " s" proper-noun arm is skipped.
func (r *SpellingCheckRule) FilterSuggestions(suggestions []string) []string {
	if r == nil || len(suggestions) == 0 {
		return suggestions
	}
	var newSuggestions []string
	for _, replacement := range suggestions {
		if r.IsProhibited(replacement) {
			continue
		}
		// Java: endsWith(" s") && isProperNoun(without last 2 chars)
		if strings.HasSuffix(replacement, " s") && len(replacement) > 3 {
			withoutS := replacement[:len(replacement)-2]
			if r.isProperNoun(withoutS) {
				// Java inserts sugg2 (Name's) then sugg1 (Name) at front via add(0,...)
				// Order after both add(0): first Name, then Name's (second add(0) pushes Name's to front... 
				// add(0, sugg1 Name); add(0, sugg2 Name's) → [Name's, Name, ...]
				newSuggestions = append([]string{withoutS + "'s", withoutS}, newSuggestions...)
				continue
			}
		}
		newSuggestions = append(newSuggestions, replacement)
	}
	newSuggestions = filterSuggestionDupes(newSuggestions)
	newSuggestions = r.filterNoSuggestWords(newSuggestions)
	if r.FilterSuggestionsExtraFn != nil {
		newSuggestions = r.FilterSuggestionsExtraFn(newSuggestions)
	}
	return newSuggestions
}

// isProperNoun ports isProperNoun: any POS tag "NNP" (English proper noun).
func (r *SpellingCheckRule) isProperNoun(word string) bool {
	if r == nil || word == "" {
		return false
	}
	if r.IsProperNounFn != nil {
		return r.IsProperNounFn(word)
	}
	if r.TagPOS == nil {
		return false
	}
	for _, t := range r.TagPOS(word) {
		if t == "NNP" {
			return true
		}
	}
	return false
}

func (r *SpellingCheckRule) filterNoSuggestWords(suggestions []string) []string {
	if r == nil || len(suggestions) == 0 {
		return suggestions
	}
	if r.FilterNoSuggestWordsFn != nil {
		return r.FilterNoSuggestWordsFn(suggestions)
	}
	return suggestions
}

func filterSuggestionDupes(in []string) []string {
	if len(in) == 0 {
		return in
	}
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}
