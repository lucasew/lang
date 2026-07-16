package ca

// CatalanMorfologikMultitokenSpeller ports
// org.languagetool.rules.ca.CatalanMorfologikMultitokenSpeller as a dict path holder.
// Full Morfologik loading is pluggable via Speller factory.
const CatalanSpellingMultitokenDict = "/ca/ca-ES_spelling_multitoken.dict"

// MultitokenSpellerFactory returns a spelling suggestions function for a dict path.
type MultitokenSpellerFactory func(dictPath string) (func(word string) []string, error)

// GetCatalanMultitokenSpellerSuggestions returns suggestions via factory, or nil if unavailable.
func GetCatalanMultitokenSpellerSuggestions(factory MultitokenSpellerFactory, word string) []string {
	if factory == nil {
		return nil
	}
	sp, err := factory(CatalanSpellingMultitokenDict)
	if err != nil || sp == nil {
		return nil
	}
	return sp(word)
}
