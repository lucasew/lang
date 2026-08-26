package de

// SwissGermanSpellerRule ports org.languagetool.rules.de.SwissGermanSpellerRule
// (extends GermanSpellerRule with CH plain-text spelling extras).
//
// Java LANGUAGE_SPECIFIC_PLAIN_TEXT_DICT = "de/hunspell/spelling-de-CH.txt"
// and init() loads ignore words from "/de/hunspell/spelling-de-CH.txt".
type SwissGermanSpellerRule struct {
	*GermanSpellerRule
}

// SwissGermanSpellingDict is Java LANGUAGE_SPECIFIC_PLAIN_TEXT_DICT (no leading slash).
const SwissGermanSpellingDict = "de/hunspell/spelling-de-CH.txt"

// SwissGermanSpellingDictResource is the classpath form used by Java init loadWords.
const SwissGermanSpellingDictResource = "/de/hunspell/spelling-de-CH.txt"

// NewSwissGermanSpellerRule ports SwissGermanSpellerRule constructors.
func NewSwissGermanSpellerRule(messages map[string]string) *SwissGermanSpellerRule {
	base := NewGermanSpellerRule(messages)
	base.LanguageVariant = "CH"
	base.LanguageSpecific = SwissGermanSpellingDict
	return &SwissGermanSpellerRule{GermanSpellerRule: base}
}

// GetID ports SwissGermanSpellerRule.getId().
func (r *SwissGermanSpellerRule) GetID() string {
	return "SWISS_GERMAN_SPELLER_RULE"
}

// GetLanguageSpecificPlainTextDict ports the Java LANGUAGE_SPECIFIC_PLAIN_TEXT_DICT field.
func (r *SwissGermanSpellerRule) GetLanguageSpecificPlainTextDict() string {
	return SwissGermanSpellingDict
}

// GetLanguageSpecificPlainTextDictResource is the resource path Java loadWords uses.
func (r *SwissGermanSpellerRule) GetLanguageSpecificPlainTextDictResource() string {
	return SwissGermanSpellingDictResource
}

// InitLanguageSpecificIgnoreWords ports SwissGermanSpellerRule.init() ignore load
// (including ß→ss rewrite and LineExpander flags via LoadIgnoreWordsFromFile).
func (r *SwissGermanSpellerRule) InitLanguageSpecificIgnoreWords(fsPath string) error {
	if r == nil || r.GermanSpellerRule == nil {
		return nil
	}
	return r.LoadIgnoreWordsFromFile(fsPath)
}

// InitFromDiscoveredResources loads DE base resources plus CH-specific files/dict.
func (r *SwissGermanSpellerRule) InitFromDiscoveredResources() error {
	if r == nil || r.GermanSpellerRule == nil {
		return nil
	}
	return r.GermanSpellerRule.InitFromDiscoveredResources()
}
