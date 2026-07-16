package ca

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/multitoken"

// CatalanMultitokenSpeller ports org.languagetool.rules.ca.CatalanMultitokenSpeller.
// Morfologik additional suggestions are a pluggable hook (nil by default).
type CatalanMultitokenSpeller struct {
	*multitoken.MultitokenSpeller
	// AdditionalSuggestions optional Morfologik-backed suggestions for originalWord.
	AdditionalSuggestions func(originalWord string) []string
}

func NewCatalanMultitokenSpeller() *CatalanMultitokenSpeller {
	return &CatalanMultitokenSpeller{MultitokenSpeller: multitoken.NewMultitokenSpeller()}
}

// Resource paths used by the Java constructor (for loaders/embed).
var CatalanMultitokenResourcePaths = []string{
	"/ca/multiwords.txt",
	"/spelling_global.txt",
	"/ca/hyphenated_words.txt",
}
