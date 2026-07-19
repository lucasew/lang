package nl

import (
	"os"
	"strings"
	"unicode/utf8"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/multitoken"
)

// DutchMultitokenSpeller ports org.languagetool.rules.nl.DutchMultitokenSpeller.
// Java loads "/nl/multiwords.txt" and "/spelling_global.txt".
type DutchMultitokenSpeller struct {
	*multitoken.MultitokenSpeller
}

// Resource paths from the Java constructor (classpath form).
var DutchMultitokenResourcePaths = []string{
	"/nl/multiwords.txt",
	"/spelling_global.txt",
}

// NewDutchMultitokenSpeller builds an empty speller with Dutch isException.
// Default Language.prepareLineForSpeller is identity (no EN/DE prepare filter).
func NewDutchMultitokenSpeller() *DutchMultitokenSpeller {
	sp := multitoken.NewMultitokenSpeller()
	d := &DutchMultitokenSpeller{MultitokenSpeller: sp}
	sp.IsException = d.IsException
	return d
}

// DutchMultitokenSpellerInstance mirrors Java INSTANCE (empty until Load*).
var DutchMultitokenSpellerInstance = NewDutchMultitokenSpeller()

// IsException ports DutchMultitokenSpeller.isException (rune-safe for ’s).
// Java uses String char indices; for BMP apostrophes rune count matches.
func (d *DutchMultitokenSpeller) IsException(original, candidate string) bool {
	if original == "" || candidate == "" {
		return false
	}
	if utf8.RuneCountInString(original) <= 2 {
		return false
	}
	runes := []rune(original)
	// original without last char equals candidate && ends with s or -
	if string(runes[:len(runes)-1]) == candidate {
		last := runes[len(runes)-1]
		if last == 's' || last == '-' {
			return true
		}
	}
	// original without last two chars equals candidate && ends with 's or ’s
	if len(runes) >= 2 && string(runes[:len(runes)-2]) == candidate {
		if strings.HasSuffix(original, "'s") || strings.HasSuffix(original, "’s") {
			return true
		}
	}
	return false
}

// LoadDutchMultitokenSpeller loads official resource files in Java constructor order.
// Empty paths are skipped.
func LoadDutchMultitokenSpeller(paths ...string) (*DutchMultitokenSpeller, error) {
	s := NewDutchMultitokenSpeller()
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

// DiscoverAndLoadDutchMultitokenSpeller finds multiwords + spelling_global and loads them.
// Missing resources → empty speller (fail-closed; no invent list).
func DiscoverAndLoadDutchMultitokenSpeller() *DutchMultitokenSpeller {
	var paths []string
	if p := spelling.DiscoverSpellingResource("nl/multiwords.txt"); p != "" {
		paths = append(paths, p)
	}
	if p := spelling.DiscoverSpellingGlobal(); p != "" {
		paths = append(paths, p)
	}
	if len(paths) == 0 {
		return NewDutchMultitokenSpeller()
	}
	s, err := LoadDutchMultitokenSpeller(paths...)
	if err != nil || s == nil {
		return NewDutchMultitokenSpeller()
	}
	return s
}

// DiscoverDutchMultiwords returns the absolute path to nl/multiwords.txt when found.
func DiscoverDutchMultiwords() string {
	return spelling.DiscoverSpellingResource("nl/multiwords.txt")
}
