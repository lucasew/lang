package en

import "strings"

// EnglishAddHyphenSuggestions ports AbstractEnglishSpellerRule.addHyphenSuggestions:
// for each misspelled hyphen part, take first suggestion for that part and rebuild
// the full hyphenated word (assumes only one part has a typo).
//
// isMisspelled / suggest must be non-nil for useful results (fail-closed empty).
func EnglishAddHyphenSuggestions(
	parts []string,
	isMisspelled func(string) bool,
	suggest func(string) []string,
) []string {
	if len(parts) == 0 || isMisspelled == nil || suggest == nil {
		return nil
	}
	var out []string
	for i, part := range parts {
		if part == "" {
			continue
		}
		if !isMisspelled(part) {
			continue
		}
		partSugs := suggest(part)
		if len(partSugs) == 0 {
			continue
		}
		out = append(out, hyphenatedWordSuggestion(parts, i, partSugs[0]))
	}
	return out
}

// hyphenatedWordSuggestion ports getHyphenatedWordSuggestion.
func hyphenatedWordSuggestion(parts []string, currentPos int, currentPartSuggestion string) string {
	newParts := make([]string, len(parts))
	for j, p := range parts {
		if j == currentPos {
			newParts[j] = currentPartSuggestion
		} else {
			newParts[j] = p
		}
	}
	return strings.Join(newParts, "-")
}
