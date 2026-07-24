package en

// MorfologikAmericanSpellerRule ports
// org.languagetool.rules.en.MorfologikAmericanSpellerRule.
//
// Java constants (see NewMorfologikAmericanSpellerRule):
//   RULE_ID = "MORFOLOGIK_RULE_EN_US"
//   RESOURCE_FILENAME = "/en/hunspell/en_US.dict"
//   LANGUAGE_SPECIFIC_PLAIN_TEXT_DICT = "en/hunspell/spelling_en-US.txt"
//   BRITISH_ENGLISH = loadWordlist("en/en-US-GB.txt", 1)
//
// Construction, isValidInOtherVariant, getFileName/getId/
// getLanguageVariantSpellingFileName, and getAdditionalTopSuggestions (automize*)
// live on MorfologikVariantSpellerRule / NewMorfologikAmericanSpellerRuleWithUser
// in morfologik_variant_speller_rules.go.
type MorfologikAmericanSpellerRule = MorfologikVariantSpellerRule
