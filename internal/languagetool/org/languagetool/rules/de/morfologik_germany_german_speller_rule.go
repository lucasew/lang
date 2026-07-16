package de

const MorfologikGermanyGermanDict = "/de/hunspell/de_DE.dict"

// MorfologikGermanyGermanSpellerRule is a non-compound DE morfologik speller stand-in.
// Prefer GermanSpellerRule; constructor: NewMorfologikGermanyGermanSpellerRule.
type MorfologikGermanyGermanSpellerRule = GermanSpellerRule

func (r *GermanSpellerRule) GetMorfologikDictFilename() string {
	return MorfologikGermanyGermanDict
}
