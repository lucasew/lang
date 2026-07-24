package en

// MorfologikCanadianSpellerRule ports
// org.languagetool.rules.en.MorfologikCanadianSpellerRule.
//
// Java constants (see NewMorfologikCanadianSpellerRule):
//   RULE_ID = "MORFOLOGIK_RULE_EN_CA"
//   RESOURCE_FILENAME = "/en/hunspell/en_CA.dict"
//   LANGUAGE_SPECIFIC_PLAIN_TEXT_DICT = "en/hunspell/spelling_en-CA.txt"
//   US_ENGLISH = loadWordlist("en/en-US-GB.txt", 0)
//
// Construction, isValidInOtherVariant → VariantInfo("American English", otherVariant),
// getFileName/getId/getLanguageVariantSpellingFileName live on
// MorfologikVariantSpellerRule / NewMorfologikCanadianSpellerRuleWithUser
// in morfologik_variant_speller_rules.go.
type MorfologikCanadianSpellerRule = MorfologikVariantSpellerRule
