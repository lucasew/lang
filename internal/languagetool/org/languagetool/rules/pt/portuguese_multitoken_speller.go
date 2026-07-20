package pt

import (
	"os"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/multitoken"
)

// PortugueseMultitokenSpeller ports org.languagetool.rules.pt.PortugueseMultitokenSpeller.
// Java loads: /pt/multiwords.txt, /spelling_global.txt, /pt/hyphenated_words.txt
// via MultitokenSpeller → language.prepareLineForSpeller on each line.
type PortugueseMultitokenSpeller struct {
	*multitoken.MultitokenSpeller
}

// PortugueseMultitokenResourcePaths ports Java constructor Arrays.asList order.
var PortugueseMultitokenResourcePaths = []string{
	"/pt/multiwords.txt",
	"/spelling_global.txt",
	"/pt/hyphenated_words.txt",
}

func NewPortugueseMultitokenSpeller() *PortugueseMultitokenSpeller {
	sp := multitoken.NewMultitokenSpeller()
	// Java MultitokenSpeller.initMultitokenSpeller → language.prepareLineForSpeller
	sp.PrepareLine = language.PortuguesePrepareLineForSpeller
	return &PortugueseMultitokenSpeller{MultitokenSpeller: sp}
}

// PortugueseMultitokenSpellerInstance mirrors Java INSTANCE.
var PortugueseMultitokenSpellerInstance = NewPortugueseMultitokenSpeller()

// LoadPortugueseMultitokenSpeller loads official resources in Java order.
func LoadPortugueseMultitokenSpeller(paths ...string) (*PortugueseMultitokenSpeller, error) {
	s := NewPortugueseMultitokenSpeller()
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

// DiscoverAndLoadPortugueseMultitokenSpeller finds multiwords + spelling_global + hyphenated.
func DiscoverAndLoadPortugueseMultitokenSpeller() *PortugueseMultitokenSpeller {
	var paths []string
	if p := spelling.DiscoverSpellingResource("pt/multiwords.txt"); p != "" {
		paths = append(paths, p)
	}
	if p := spelling.DiscoverSpellingGlobal(); p != "" {
		paths = append(paths, p)
	}
	if p := spelling.DiscoverSpellingResource("pt/hyphenated_words.txt"); p != "" {
		paths = append(paths, p)
	}
	if len(paths) == 0 {
		return NewPortugueseMultitokenSpeller()
	}
	s, err := LoadPortugueseMultitokenSpeller(paths...)
	if err != nil || s == nil {
		return NewPortugueseMultitokenSpeller()
	}
	return s
}
