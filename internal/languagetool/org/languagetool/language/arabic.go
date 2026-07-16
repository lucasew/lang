package language

// Arabic ports org.languagetool.language.Arabic metadata and factory surface.
type ArabicLang struct {
	ShortCode, Name, SpellerRuleID string
	Countries                      []string
	Maintained                     string
}

func (a ArabicLang) GetName() string      { return a.Name }
func (a ArabicLang) GetShortCode() string { return a.ShortCode }
func (a ArabicLang) GetCountries() []string {
	out := make([]string, len(a.Countries))
	copy(out, a.Countries)
	return out
}

// Arabic is the singleton language descriptor.
var Arabic = ArabicLang{
	ShortCode:      "ar",
	Name:           "Arabic",
	SpellerRuleID:  "HUNSPELL_RULE_AR",
	Maintained:     "ActivelyMaintained",
	Countries: []string{
		"", "SA", "DZ", "BH", "EG", "IQ", "JO", "KW", "LB", "LY",
		"MA", "OM", "QA", "SD", "SY", "TN", "AE", "YE",
	},
}

// ArabicRelevantRuleIDs lists rule IDs registered by Arabic.getRelevantRules (names only).
func ArabicRelevantRuleIDs() []string {
	return []string{
		"MULTIPLE_WHITESPACE",
		"SENTENCE_WHITESPACE",
		"UNPAIRED_BRACKETS",
		"COMMA_WHITESPACE",
		"LONG_SENTENCE",
		"HUNSPELL_RULE_AR",
		"ARABIC_COMMA_WHITESPACE",
		"ARABIC_QUESTION_MARK_WHITESPACE",
		"ARABIC_SEMICOLON_WHITESPACE",
		"ARABIC_DOUBLE_PUNCTUATION",
		"ARABIC_WORD_REPEAT",
		"ARABIC_SIMPLE_REPLACE",
		"ARABIC_DIACRITICS",
		"ARABIC_DARJA",
		"ARABIC_HOMOPHONES",
		"ARABIC_REDUNDANCY",
		"ARABIC_WORD_COHERENCY",
		"ARABIC_WORDINESS",
		"ARABIC_WRONG_WORD_IN_CONTEXT",
		"ARABIC_TRANS_VERB",
		"ARABIC_INFLECTED_ONE_WORD_REPLACE",
	}
}

func ArabicMaintainers() []string {
	return []string{"Taha Zerrouki", "Sohaib Afifi"}
}
