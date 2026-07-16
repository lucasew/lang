package language

import "regexp"

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
