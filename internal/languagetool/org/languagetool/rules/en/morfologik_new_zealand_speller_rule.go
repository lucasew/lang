package en

// MorfologikNewZealandSpellerRule ports
// org.languagetool.rules.en.MorfologikNewZealandSpellerRule.
//
// Java constants (see NewMorfologikNewZealandSpellerRule):
//   RULE_ID = "MORFOLOGIK_RULE_EN_NZ"
//   RESOURCE_FILENAME = "/en/hunspell/en_NZ.dict"
//   LANGUAGE_SPECIFIC_PLAIN_TEXT_DICT = "en/hunspell/spelling_en-NZ.txt"
//   US_ENGLISH = loadWordlist("en/en-US-GB.txt", 0)
//
// Construction, isValidInOtherVariant → VariantInfo("American English", otherVariant),
// getFileName/getId/getLanguageVariantSpellingFileName live on
// MorfologikVariantSpellerRule / NewMorfologikNewZealandSpellerRuleWithUser
// in morfologik_variant_speller_rules.go.
type MorfologikNewZealandSpellerRule = MorfologikVariantSpellerRule
