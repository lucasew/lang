package en

import "regexp"

// Java CONTAINS_TOKEN — drop "timezone s" style splits from suggestions.
var englishContainsToken = regexp.MustCompile(`.* (b|c|d|e|f|g|h|j|k|l|m|n|o|p|q|r|s|t|v|w|y|z|ll|ve)$`)

// filterEnglishContainsToken ports AbstractEnglishSpellerRule.filterSuggestions extra arm.
func filterEnglishContainsToken(suggestions []string) []string {
	if len(suggestions) == 0 {
		return suggestions
	}
	out := make([]string, 0, len(suggestions))
	for _, s := range suggestions {
		if englishContainsToken.MatchString(s) {
			continue
		}
		out = append(out, s)
	}
	return out
}
