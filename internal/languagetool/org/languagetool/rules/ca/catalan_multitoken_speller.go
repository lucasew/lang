package ca

import (
	"os"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/multitoken"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// CatalanMultitokenSpeller ports org.languagetool.rules.ca.CatalanMultitokenSpeller.
// Java loads: /ca/multiwords.txt, /spelling_global.txt, /ca/hyphenated_words.txt
// and wires CatalanMorfologikMultitokenSpeller.getSpeller() into getAdditionalSuggestions.
type CatalanMultitokenSpeller struct {
	*multitoken.MultitokenSpeller
}

// CatalanMultitokenResourcePaths ports Java constructor Arrays.asList order.
var CatalanMultitokenResourcePaths = []string{
	"/ca/multiwords.txt",
	"/spelling_global.txt",
	"/ca/hyphenated_words.txt",
}

// NewCatalanMultitokenSpeller builds speller with Morfologik additional suggestions hook.
func NewCatalanMultitokenSpeller() *CatalanMultitokenSpeller {
	sp := multitoken.NewMultitokenSpeller()
	c := &CatalanMultitokenSpeller{MultitokenSpeller: sp}
	// Java MultitokenSpeller.initMultitokenSpeller → language.prepareLineForSpeller
	sp.PrepareLine = language.CatalanPrepareLineForSpeller
	// Java: this.speller = CatalanMorfologikMultitokenSpeller.getSpeller();
	// getAdditionalSuggestions → speller.getSuggestions
	sp.GetAdditionalSuggestions = func(originalWord string) []multitoken.WeightedSuggestion {
		return GetWeightedSuggestions(originalWord)
	}
	return c
}

// CatalanMultitokenSpellerInstance mirrors Java INSTANCE.
var CatalanMultitokenSpellerInstance = NewCatalanMultitokenSpeller()

// LoadCatalanMultitokenSpeller loads official resources in Java order.
func LoadCatalanMultitokenSpeller(paths ...string) (*CatalanMultitokenSpeller, error) {
	s := NewCatalanMultitokenSpeller()
	for _, p := range paths {
		if tools.JavaStringTrim(p) == "" {
			continue
		}
		f, err := os.Open(p)
		if err != nil {
			return nil, err
		}
		err = s.LoadWords(f)
		_ = f.Close()
		if err != nil {
			return nil, err
		}
	}
	return s, nil
}

// DiscoverAndLoadCatalanMultitokenSpeller finds multiwords + spelling_global + hyphenated.
func DiscoverAndLoadCatalanMultitokenSpeller() *CatalanMultitokenSpeller {
	var paths []string
	for _, rel := range []string{"ca/multiwords.txt", "ca/hyphenated_words.txt"} {
		if p := spelling.DiscoverSpellingResource(rel); p != "" {
			paths = append(paths, p)
		}
	}
	if p := spelling.DiscoverSpellingGlobal(); p != "" {
		// insert global after multiwords like Java order: multiwords, global, hyphenated
		// rebuild: multiwords, global, hyphenated
		ordered := []string{}
		if p0 := spelling.DiscoverSpellingResource("ca/multiwords.txt"); p0 != "" {
			ordered = append(ordered, p0)
		}
		ordered = append(ordered, p)
		if p2 := spelling.DiscoverSpellingResource("ca/hyphenated_words.txt"); p2 != "" {
			ordered = append(ordered, p2)
		}
		paths = ordered
	}
	if len(paths) == 0 {
		return NewCatalanMultitokenSpeller()
	}
	s, err := LoadCatalanMultitokenSpeller(paths...)
	if err != nil || s == nil {
		return NewCatalanMultitokenSpeller()
	}
	return s
}
