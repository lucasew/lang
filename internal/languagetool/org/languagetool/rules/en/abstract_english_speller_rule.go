package en

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/morfologik"
)

// English spelling resource paths used by AbstractEnglishSpellerRule.
const (
	EnglishCustomSpellingFile = "/en/hunspell/spelling.txt" // language-relative custom
	EnglishGlobalSpellingFile = "/spelling.txt"
	EnglishMultiwordsFile     = "/en/multiwords.txt"
)

// EnglishDoNotSuggestWords ports the NOSUGGEST / blocked list (subset).
var EnglishDoNotSuggestWords = map[string]struct{}{
	"asshole": {}, "assholes": {}, "bullshit": {}, "cunt": {},
	"germane": {}, "double check": {}, "flat screen": {},
	"full time": {}, "part time": {}, "java script": {}, "off topic": {},
	"hard coding": {}, "hard coded": {}, "fine tune": {}, "fine tuned": {},
	"fine tuning": {}, "hands free": {},
}

// AbstractEnglishSpellerRule ports
// org.languagetool.rules.en.AbstractEnglishSpellerRule surface over MorfologikSpellerRule.
type AbstractEnglishSpellerRule struct {
	*morfologik.MorfologikSpellerRule
	LanguageShortCode string // e.g. "en"
	// VariantCode for file paths (en-US etc.).
	VariantCode string
}

func NewAbstractEnglishSpellerRule(id, variantCode string, speller *morfologik.MorfologikSpeller) *AbstractEnglishSpellerRule {
	short := "en"
	if i := strings.Index(variantCode, "-"); i > 0 {
		short = variantCode[:i]
	} else if variantCode != "" {
		short = variantCode
	}
	base := morfologik.NewMorfologikSpellerRule(id, short, "", speller)
	return &AbstractEnglishSpellerRule{
		MorfologikSpellerRule: base,
		LanguageShortCode:     short,
		VariantCode:           variantCode,
	}
}

// GetAdditionalSpellingFileNames ports getAdditionalSpellingFileNames.
func (r *AbstractEnglishSpellerRule) GetAdditionalSpellingFileNames() []string {
	custom := r.LanguageShortCode + "/hunspell/spelling.txt"
	if r.VariantCode != "" {
		// Java: language.getShortCode() + CUSTOM_SPELLING_FILE
		custom = r.LanguageShortCode + "/hunspell/spelling.txt"
	}
	return []string{custom, EnglishGlobalSpellingFile, EnglishMultiwordsFile}
}

// IsDoNotSuggest reports whether the word is blocked from suggestions.
func IsDoNotSuggest(word string) bool {
	_, ok := EnglishDoNotSuggestWords[strings.ToLower(strings.TrimSpace(word))]
	return ok
}

// FilterSuggestions drops blocked suggestions.
func FilterEnglishSuggestions(suggestions []string) []string {
	var out []string
	for _, s := range suggestions {
		if IsDoNotSuggest(s) {
			continue
		}
		out = append(out, s)
	}
	return out
}
