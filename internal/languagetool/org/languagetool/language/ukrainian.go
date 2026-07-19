package language

import (
	"regexp"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// UkrainianIgnoredChars ports Ukrainian.IGNORED_CHARS.
var UkrainianIgnoredChars = regexp.MustCompile("[\u00AD\u0301]")

// UkrainianLang is Ukrainian language metadata (distinct from SmallLang Ukrainian).
type UkrainianLanguage struct {
	ShortCode, Name, SpellerRuleID string
	Countries                      []string
	RuleFiles                      []string
}

func (u UkrainianLanguage) GetName() string      { return u.Name }
func (u UkrainianLanguage) GetShortCode() string { return u.ShortCode }
func (u UkrainianLanguage) GetCountries() []string {
	out := make([]string, len(u.Countries))
	copy(out, u.Countries)
	return out
}

// GetShortCodeWithCountryAndVariant ports Language.buildShortCodeWithCountryAndVariant.
func (u UkrainianLanguage) GetShortCodeWithCountryAndVariant() string {
	return BuildShortCodeWithCountryAndVariant(u.ShortCode, u.Countries, "")
}

// GetCommonWordsPath ports Language.getCommonWordsPath → uk/common_words.txt.
func (u UkrainianLanguage) GetCommonWordsPath() string {
	return DefaultCommonWordsPath(u.GetShortCode())
}

// GetIgnoredCharactersRegex ports Ukrainian.getIgnoredCharactersRegex.
func (u UkrainianLanguage) GetIgnoredCharactersRegex() *regexp.Regexp {
	return languagetool.UkrainianIgnoredCharactersRegex
}

// GetDefaultSpellingRuleID ports MorfologikUkrainianSpellerRule getId.
func (u UkrainianLanguage) GetDefaultSpellingRuleID() string {
	return u.SpellerRuleID
}

// GetOpeningDoubleQuote ports Ukrainian.getOpeningDoubleQuote ("«").
func (u UkrainianLanguage) GetOpeningDoubleQuote() string { return "«" }

// GetClosingDoubleQuote ports Ukrainian.getClosingDoubleQuote ("»").
func (u UkrainianLanguage) GetClosingDoubleQuote() string { return "»" }

// GetOpeningSingleQuote ports Ukrainian.getOpeningSingleQuote ("‘").
func (u UkrainianLanguage) GetOpeningSingleQuote() string { return "‘" }

// GetClosingSingleQuote ports Ukrainian.getClosingSingleQuote ("’").
func (u UkrainianLanguage) GetClosingSingleQuote() string { return "’" }

// IsAdvancedTypographyEnabled ports Ukrainian.isAdvancedTypographyEnabled (false).
func (u UkrainianLanguage) IsAdvancedTypographyEnabled() bool { return false }

// ToAdvancedTypography ports Ukrainian.toAdvancedTypography (disabled → suggestion tags only).
func (u UkrainianLanguage) ToAdvancedTypography(input string) string {
	return UkrainianAdvancedTypography(input)
}

// UkrainianLanguageDefault is the default Ukrainian variant descriptor.
var UkrainianLanguageDefault = UkrainianLanguage{
	ShortCode:     "uk",
	Name:          "Ukrainian",
	SpellerRuleID: "MORFOLOGIK_RULE_UK_UA",
	Countries:     []string{"UA"},
	RuleFiles: []string{
		"grammar-spelling.xml",
		"grammar-grammar.xml",
		"grammar-barbarism.xml",
		"grammar-style.xml",
		"grammar-punctuation.xml",
	},
}

// GetRuleFileNames ports Ukrainian.getRuleFileNames — base grammar.xml then RULE_FILES.
func (u UkrainianLanguage) GetRuleFileNames() []string {
	const dirBase = "/org/languagetool/rules/uk/"
	out := []string{dirBase + "grammar.xml"}
	for _, f := range u.RuleFiles {
		out = append(out, dirBase+f)
	}
	return out
}

// GetMaintainedState ports Ukrainian.getMaintainedState → ActivelyMaintained.
func (u UkrainianLanguage) GetMaintainedState() languagetool.LanguageMaintainedState {
	return languagetool.ActivelyMaintained
}

// GetMaintainers ports Ukrainian.getMaintainers (Andriy Rysin, Maksym Davydov).
func (u UkrainianLanguage) GetMaintainers() []Contributor {
	return []Contributor{
		NewContributor("Andriy Rysin"),
		NewContributor("Maksym Davydov"),
	}
}

// GetRelevantRuleIDs ports Ukrainian.getRelevantRules IDs.
func (u UkrainianLanguage) GetRelevantRuleIDs() []string {
	return UkrainianRelevantRuleIDs()
}
