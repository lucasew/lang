package de

// AustrianGermanSpellerRule ports org.languagetool.rules.de.AustrianGermanSpellerRule
// (extends GermanSpellerRule with AT plain-text spelling extras).
//
// Java LANGUAGE_SPECIFIC_PLAIN_TEXT_DICT = "de/hunspell/spelling-de-AT.txt"
// and init() loads ignore words from "/de/hunspell/spelling-de-AT.txt".
type AustrianGermanSpellerRule struct {
	*GermanSpellerRule
}

// AustrianGermanSpellingDict is Java LANGUAGE_SPECIFIC_PLAIN_TEXT_DICT (no leading slash).
const AustrianGermanSpellingDict = "de/hunspell/spelling-de-AT.txt"

// AustrianGermanSpellingDictResource is the classpath form used by Java init loadWords.
const AustrianGermanSpellingDictResource = "/de/hunspell/spelling-de-AT.txt"

// NewAustrianGermanSpellerRule ports AustrianGermanSpellerRule constructors.
func NewAustrianGermanSpellerRule(messages map[string]string) *AustrianGermanSpellerRule {
	base := NewGermanSpellerRule(messages)
	base.LanguageVariant = "AT"
	base.LanguageSpecific = AustrianGermanSpellingDict
	return &AustrianGermanSpellerRule{GermanSpellerRule: base}
}

// GetID ports AustrianGermanSpellerRule.getId().
func (r *AustrianGermanSpellerRule) GetID() string {
	return "AUSTRIAN_GERMAN_SPELLER_RULE"
}

// GetLanguageSpecificPlainTextDict ports the Java LANGUAGE_SPECIFIC_PLAIN_TEXT_DICT field.
func (r *AustrianGermanSpellerRule) GetLanguageSpecificPlainTextDict() string {
	return AustrianGermanSpellingDict
}

// GetLanguageSpecificPlainTextDictResource is the resource path Java loadWords uses.
func (r *AustrianGermanSpellerRule) GetLanguageSpecificPlainTextDictResource() string {
	return AustrianGermanSpellingDictResource
}

// InitLanguageSpecificIgnoreWords ports AustrianGermanSpellerRule.init() ignore load
// from an on-disk path resolving the official resource (host supplies filesystem path).
func (r *AustrianGermanSpellerRule) InitLanguageSpecificIgnoreWords(fsPath string) error {
	if r == nil || r.GermanSpellerRule == nil {
		return nil
	}
	return r.LoadIgnoreWordsFromFile(fsPath)
}

// InitFromDiscoveredResources loads DE base resources plus AT-specific files/dict.
func (r *AustrianGermanSpellerRule) InitFromDiscoveredResources() error {
	if r == nil || r.GermanSpellerRule == nil {
		return nil
	}
	return r.GermanSpellerRule.InitFromDiscoveredResources()
}
