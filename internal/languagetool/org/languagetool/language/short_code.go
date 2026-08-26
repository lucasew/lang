package language

import "strings"

// BuildShortCodeWithCountryAndVariant ports Language.buildShortCodeWithCountryAndVariant.
// shortCode is getShortCode(); countries is getCountries(); variant is getVariant() or "".
// Private-use tags containing "-x-" are returned unchanged (e.g. de-DE-x-simple-language).
func BuildShortCodeWithCountryAndVariant(shortCode string, countries []string, variant string) string {
	if strings.Contains(shortCode, "-x-") {
		return shortCode
	}
	name := shortCode
	if len(countries) == 1 {
		name += "-" + countries[0]
		if variant != "" {
			name += "-" + variant
		}
	}
	return name
}

// DefaultCommonWordsPath ports Language.getCommonWordsPath default: shortCode/common_words.txt.
func DefaultCommonWordsPath(shortCode string) string {
	return shortCode + "/common_words.txt"
}

// CommonWordsPathNone marks Java overrides that return null (no common words file).
// Callers treat empty string as nil path.
const CommonWordsPathNone = ""
