package en

import (
	"os"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/multitoken"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// EnglishMultitokenSpeller ports org.languagetool.rules.en.EnglishMultitokenSpeller.
// Java:
//
//	super(Languages.getLanguageForShortCode("en"),
//	  Arrays.asList("/en/multiwords.txt", "/spelling_global.txt"));
//
// MultitokenSpeller.initMultitokenSpeller loads those classpath resources and
// applies language.prepareLineForSpeller (English: NN*/JJ* forms only; '+' → empty).
type EnglishMultitokenSpeller struct {
	*multitoken.MultitokenSpeller
}

// EnglishMultitokenResourcePaths ports Java constructor Arrays.asList order.
var EnglishMultitokenResourcePaths = []string{
	"/en/multiwords.txt",
	"/spelling_global.txt",
}

// NewEnglishMultitokenSpeller builds an empty speller with English prepareLineForSpeller.
// Java MultitokenSpeller ctor always calls initMultitokenSpeller; use INSTANCE or
// DiscoverAndLoadEnglishMultitokenSpeller / LoadEnglishMultitokenSpeller for loaded twins.
func NewEnglishMultitokenSpeller() *EnglishMultitokenSpeller {
	sp := multitoken.NewMultitokenSpeller()
	// Java MultitokenSpeller.initMultitokenSpeller → language.prepareLineForSpeller
	sp.PrepareLine = language.EnglishPrepareLineForSpeller
	return &EnglishMultitokenSpeller{MultitokenSpeller: sp}
}

// INSTANCE ports EnglishMultitokenSpeller.INSTANCE.
// Java static construction loads /en/multiwords.txt + /spelling_global.txt via DataBroker.
// Go discovers the same official files under the inspiration tree (no invent wordlists);
// if none are found, INSTANCE is an empty prepared speller (fail-closed).
var INSTANCE = DiscoverAndLoadEnglishMultitokenSpeller()

// EnglishMultitokenSpellerInstance is the same singleton under the sibling-lang naming style.
var EnglishMultitokenSpellerInstance = INSTANCE

// PrepareLineForSpeller ports English.prepareLineForSpeller.
// multiwords.txt lines are form\tPOS; only NN*/JJ* forms enter the multitoken dict.
// Delegates to language.EnglishPrepareLineForSpeller (single Java twin implementation).
func PrepareLineForSpeller(line string) []string {
	return language.EnglishPrepareLineForSpeller(line)
}

// LoadEnglishMultitokenSpeller loads official resource files in Java constructor order.
// Empty paths are skipped. Paths are filesystem paths to the same contents Java serves
// from classpath /en/multiwords.txt and /spelling_global.txt.
func LoadEnglishMultitokenSpeller(paths ...string) (*EnglishMultitokenSpeller, error) {
	s := NewEnglishMultitokenSpeller()
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

// DiscoverAndLoadEnglishMultitokenSpeller finds official multiwords + spelling_global
// (Java order: /en/multiwords.txt then /spelling_global.txt) and loads them.
// Missing resources → empty speller (fail-closed; no invent list).
func DiscoverAndLoadEnglishMultitokenSpeller() *EnglishMultitokenSpeller {
	var paths []string
	if p := spelling.DiscoverSpellingResource("en/multiwords.txt"); p != "" {
		paths = append(paths, p)
	}
	if p := spelling.DiscoverSpellingGlobal(); p != "" {
		paths = append(paths, p)
	}
	if len(paths) == 0 {
		return NewEnglishMultitokenSpeller()
	}
	s, err := LoadEnglishMultitokenSpeller(paths...)
	if err != nil || s == nil {
		return NewEnglishMultitokenSpeller()
	}
	return s
}

// DiscoverEnglishMultiwords returns the absolute path to en/multiwords.txt when found
// (Java resource /en/multiwords.txt used by EnglishMultitokenSpeller).
func DiscoverEnglishMultiwords() string {
	return spelling.DiscoverSpellingResource("en/multiwords.txt")
}
