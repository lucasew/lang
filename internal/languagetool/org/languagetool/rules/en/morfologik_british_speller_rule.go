package en

// MorfologikBritishSpellerRule ports
// org.languagetool.rules.en.MorfologikBritishSpellerRule.
//
// Java constants (see NewMorfologikBritishSpellerRule):
//   RULE_ID = "MORFOLOGIK_RULE_EN_GB"
//   RESOURCE_FILENAME = "/en/hunspell/en_GB.dict"
//   LANGUAGE_SPECIFIC_PLAIN_TEXT_DICT = "en/hunspell/spelling_en-GB.txt"
//   US_ENGLISH = loadWordlist("en/en-US-GB.txt", 0)
//
// Construction, isValidInOtherVariant → VariantInfo("American English", otherVariant),
// getFileName/getId/getLanguageVariantSpellingFileName live on
// MorfologikVariantSpellerRule / NewMorfologikBritishSpellerRuleWithUser
// in morfologik_variant_speller_rules.go.
type MorfologikBritishSpellerRule = MorfologikVariantSpellerRule
