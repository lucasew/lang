package fr

import (
	"os"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/multitoken"
)

// FrenchMultitokenSpeller ports org.languagetool.rules.fr.FrenchMultitokenSpeller.
// Java loads: /fr/multiwords.txt, /spelling_global.txt, /fr/hyphenated_words.txt
type FrenchMultitokenSpeller struct {
	*multitoken.MultitokenSpeller
}

// FrenchMultitokenResourcePaths ports Java constructor Arrays.asList order.
var FrenchMultitokenResourcePaths = []string{
	"/fr/multiwords.txt",
	"/spelling_global.txt",
	"/fr/hyphenated_words.txt",
}

func NewFrenchMultitokenSpeller() *FrenchMultitokenSpeller {
	return &FrenchMultitokenSpeller{MultitokenSpeller: multitoken.NewMultitokenSpeller()}
}

// FrenchMultitokenSpellerInstance mirrors Java INSTANCE.
var FrenchMultitokenSpellerInstance = NewFrenchMultitokenSpeller()

// LoadFrenchMultitokenSpeller loads official resources in Java order.
func LoadFrenchMultitokenSpeller(paths ...string) (*FrenchMultitokenSpeller, error) {
	s := NewFrenchMultitokenSpeller()
	for _, p := range paths {
		if strings.TrimSpace(p) == "" {
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

// DiscoverAndLoadFrenchMultitokenSpeller finds multiwords + spelling_global + hyphenated.
func DiscoverAndLoadFrenchMultitokenSpeller() *FrenchMultitokenSpeller {
	var paths []string
	if p := spelling.DiscoverSpellingResource("fr/multiwords.txt"); p != "" {
		paths = append(paths, p)
	}
	if p := spelling.DiscoverSpellingGlobal(); p != "" {
		paths = append(paths, p)
	}
	if p := spelling.DiscoverSpellingResource("fr/hyphenated_words.txt"); p != "" {
		paths = append(paths, p)
	}
	if len(paths) == 0 {
		return NewFrenchMultitokenSpeller()
	}
	s, err := LoadFrenchMultitokenSpeller(paths...)
	if err != nil || s == nil {
		return NewFrenchMultitokenSpeller()
	}
	return s
}
