package es

import (
	"os"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/language"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/multitoken"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"
)

// SpanishMultitokenSpeller ports org.languagetool.rules.es.SpanishMultitokenSpeller.
// Java loads: /es/multiwords.txt, /spelling_global.txt, /es/hyphenated_words.txt
// via MultitokenSpeller → language.prepareLineForSpeller on each line.
type SpanishMultitokenSpeller struct {
	*multitoken.MultitokenSpeller
}

// SpanishMultitokenResourcePaths ports Java constructor Arrays.asList order.
var SpanishMultitokenResourcePaths = []string{
	"/es/multiwords.txt",
	"/spelling_global.txt",
	"/es/hyphenated_words.txt",
}

func NewSpanishMultitokenSpeller() *SpanishMultitokenSpeller {
	sp := multitoken.NewMultitokenSpeller()
	// Java MultitokenSpeller.initMultitokenSpeller → language.prepareLineForSpeller
	sp.PrepareLine = language.SpanishPrepareLineForSpeller
	return &SpanishMultitokenSpeller{MultitokenSpeller: sp}
}

// SpanishMultitokenSpellerInstance mirrors Java INSTANCE.
var SpanishMultitokenSpellerInstance = NewSpanishMultitokenSpeller()

// LoadSpanishMultitokenSpeller loads official resources in Java order.
func LoadSpanishMultitokenSpeller(paths ...string) (*SpanishMultitokenSpeller, error) {
	s := NewSpanishMultitokenSpeller()
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

// DiscoverAndLoadSpanishMultitokenSpeller finds multiwords + spelling_global + hyphenated.
func DiscoverAndLoadSpanishMultitokenSpeller() *SpanishMultitokenSpeller {
	var paths []string
	if p := spelling.DiscoverSpellingResource("es/multiwords.txt"); p != "" {
		paths = append(paths, p)
	}
	if p := spelling.DiscoverSpellingGlobal(); p != "" {
		paths = append(paths, p)
	}
	if p := spelling.DiscoverSpellingResource("es/hyphenated_words.txt"); p != "" {
		paths = append(paths, p)
	}
	if len(paths) == 0 {
		return NewSpanishMultitokenSpeller()
	}
	s, err := LoadSpanishMultitokenSpeller(paths...)
	if err != nil || s == nil {
		return NewSpanishMultitokenSpeller()
	}
	return s
}
