package en

// MorfologikAustralianSpellerRule ports
// org.languagetool.rules.en.MorfologikAustralianSpellerRule.
//
// Java constants (see NewMorfologikAustralianSpellerRule):
//   RULE_ID = "MORFOLOGIK_RULE_EN_AU"
//   RESOURCE_FILENAME = "/en/hunspell/en_AU.dict"
//   LANGUAGE_SPECIFIC_PLAIN_TEXT_DICT = "en/hunspell/spelling_en-AU.txt"
//   US_ENGLISH = loadWordlist("en/en-US-GB.txt", 0)
//
// Construction, isValidInOtherVariant → VariantInfo("American English", otherVariant),
// getFileName/getId/getLanguageVariantSpellingFileName live on
// MorfologikVariantSpellerRule / NewMorfologikAustralianSpellerRuleWithUser
// in morfologik_variant_speller_rules.go.
type MorfologikAustralianSpellerRule = MorfologikVariantSpellerRule
