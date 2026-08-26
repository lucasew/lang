package uk

import "regexp"

// InflectionHelper is the Java-name twin for inflection utilities.
type InflectionHelper struct{}

func (InflectionHelper) GetAdjCaseInflections(posTags []string) []Inflection {
	return GetAdjCaseInflections(posTags)
}
func (InflectionHelper) GetNounCaseInflections(posTags []string) []Inflection {
	return GetNounCaseInflections(posTags)
}
func (InflectionHelper) GetNumrCaseInflections(posTags []string) []Inflection {
	return GetNumrCaseInflections(posTags)
}
func (InflectionHelper) Intersect(master, slave []Inflection) bool {
	return InflectionsIntersect(master, slave)
}
func (InflectionHelper) GetAdjInflectionsFromTags(posTags []string, postagStart string) []Inflection {
	return GetAdjInflectionsFromTags(posTags, postagStart)
}
func (InflectionHelper) GetNounInflectionsFromTags(posTags []string, ignoreRE *regexp.Regexp) []Inflection {
	return GetNounInflectionsFromTags(posTags, ignoreRE)
}

// LemmaHelper is the Java-name twin for lemma set helpers.
type LemmaHelper struct{}

func (LemmaHelper) HasLemma(lemmas []string, want map[string]struct{}) bool {
	return HasLemma(lemmas, want)
}
func (LemmaHelper) HasLemmaString(lemmas []string, want string) bool {
	return HasLemmaString(lemmas, want)
}
func (LemmaHelper) CleanIgnoreChars(token string) string { return CleanIgnoreChars(token) }
func (LemmaHelper) IsTimePlusLemma(lemma string) bool    { return IsTimePlusLemma(lemma) }

// SearchHelper is the Java-name twin for token-sequence search.
type SearchHelper struct{}

func (SearchHelper) NewMatch(tokenLine string) *SearchMatch { return NewSearchMatch(tokenLine) }

// VerbInflectionHelper is the Java-name twin for verb agreement slots.
type VerbInflectionHelper struct{}

func (VerbInflectionHelper) GetVerbInflections(posTags []string) []VerbInflection {
	return GetVerbInflections(posTags)
}
func (VerbInflectionHelper) GetNounInflections(posTags []string) []VerbInflection {
	return GetNounInflections(posTags)
}
func (VerbInflectionHelper) Overlap(verbTags, nounTags []string) bool {
	return VerbInflectionsOverlap(verbTags, nounTags)
}
