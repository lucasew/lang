package language

import "github.com/lucasew/lang/internal/languagetool/org/languagetool"

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

// GetMaintainedState ports Arabic.getMaintainedState → ActivelyMaintained.
func (a ArabicLang) GetMaintainedState() languagetool.LanguageMaintainedState {
	return languagetool.ActivelyMaintained
}

// GetMaintainers ports Arabic.getMaintainers (Taha Zerrouki, Sohaib Afifi).
func (a ArabicLang) GetMaintainers() []Contributor {
	return []Contributor{
		NewContributor("Taha Zerrouki"),
		NewContributor("Sohaib Afifi"),
	}
}

// GetShortCodeWithCountryAndVariant ports Language.buildShortCodeWithCountryAndVariant.
// Java Arabic lists many countries (len != 1) → bare "ar".
func (a ArabicLang) GetShortCodeWithCountryAndVariant() string {
	return BuildShortCodeWithCountryAndVariant(a.ShortCode, a.Countries, "")
}

// GetCommonWordsPath ports Language.getCommonWordsPath → ar/common_words.txt.
func (a ArabicLang) GetCommonWordsPath() string {
	return DefaultCommonWordsPath(a.GetShortCode())
}

// GetDefaultSpellingRuleID ports createDefaultSpellingRule id.
func (a ArabicLang) GetDefaultSpellingRuleID() string {
	return a.SpellerRuleID
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

// ArabicRelevantRuleIDs ports Arabic.getRelevantRules rule IDs (class getId / RULE_ID only).
// Order matches Java Arrays.asList in Arabic.java.
func ArabicRelevantRuleIDs() []string {
	return []string{
		"WHITESPACE_RULE",              // MultipleWhitespaceRule
		"SENTENCE_WHITESPACE",          // SentenceWhitespaceRule
		"UNPAIRED_BRACKETS",            // GenericUnpairedBracketsRule default id
		"COMMA_PARENTHESIS_WHITESPACE", // CommaWhitespaceRule
		"TOO_LONG_SENTENCE",            // LongSentenceRule.RULE_ID
		"HUNSPELL_RULE_AR",             // ArabicHunspellSpellerRule.RULE_ID
		"ARABIC_COMMA_PARENTHESIS_WHITESPACE",
		"ARABIC_QM_WHITESPACE",
		"ARABIC_SC_WHITESPACE",
		"ARABIC_DOUBLE_PUNCTUATION",
		"ARABIC_WORD_REPEAT_RULE",
		"AR_SIMPLE_REPLACE",
		"AR_DIACRITICS_REPLACE",
		"AR_DARJA_REPLACE",
		"AR_HOMOPHONES_REPLACE",
		"AR_REDUNDANCY_REPLACE",
		"AR_WORD_COHERENCY",
		"AR_WORDINESS_REPLACE",
		"ARABIC_WRONG_WORD_IN_CONTEXT",
		// Java constant AR_VERB_TRANSITIVE_IINDIRECT (typo "IINDIRECT" in upstream source)
		"AR_VERB_TRANSITIVE_IINDIRECT",
		"AR_INFLECTED_ONE_WORD",
	}
}

// GetRelevantRuleIDs ports Arabic.getRelevantRules IDs.
func (a ArabicLang) GetRelevantRuleIDs() []string {
	return ArabicRelevantRuleIDs()
}
