package spelling

// LanguageTool brand constants from SpellingCheckRule.
const (
	LanguageToolBrand   = "LanguageTool"
	LanguageToolerBrand = "LanguageTooler"
)

// AdditionalTopSuggestions ports SpellingCheckRule.getAdditionalTopSuggestions
// (string form): curated LanguageTool / LanguageTooler when missing from list.
func AdditionalTopSuggestions(existing []string, word string) []string {
	var more []string
	has := func(s string) bool {
		for _, e := range existing {
			if e == s {
				return true
			}
		}
		return false
	}
	if (word == "Languagetool" || word == "languagetool") && !has(LanguageToolBrand) {
		more = append(more, LanguageToolBrand)
	}
	if (word == "Languagetooler" || word == "languagetooler") && !has(LanguageToolerBrand) {
		more = append(more, LanguageToolerBrand)
	}
	return more
}
