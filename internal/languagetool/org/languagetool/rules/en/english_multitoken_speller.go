package en

import (
	"os"
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/rules/spelling/multitoken"
)

// EnglishMultitokenSpeller ports org.languagetool.rules.en.EnglishMultitokenSpeller.
// Java loads "/en/multiwords.txt" and "/spelling_global.txt" via MultitokenSpeller.
type EnglishMultitokenSpeller struct {
	*multitoken.MultitokenSpeller
}

// NewEnglishMultitokenSpeller builds an empty speller with English prepareLineForSpeller.
func NewEnglishMultitokenSpeller() *EnglishMultitokenSpeller {
	sp := multitoken.NewMultitokenSpeller()
	sp.PrepareLine = PrepareLineForSpeller
	return &EnglishMultitokenSpeller{MultitokenSpeller: sp}
}

// PrepareLineForSpeller ports English.prepareLineForSpeller.
// multiwords.txt lines are form\tPOS; only NN*/JJ* forms enter the multitoken dict.
func PrepareLineForSpeller(line string) []string {
	parts := strings.Split(line, "#")
	if len(parts) == 0 {
		return []string{line}
	}
	if strings.Contains(line, "+") {
		// while the morfologik separator is "+", multiwords with '+' can cause undesired results.
		return []string{""}
	}
	formTag := strings.Split(parts[0], "\t")
	form := strings.TrimSpace(formTag[0])
	if len(formTag) > 1 {
		tag := strings.TrimSpace(formTag[1])
		if strings.HasPrefix(tag, "NN") || strings.HasPrefix(tag, "JJ") {
			return []string{form}
		}
		return []string{""}
	}
	return []string{line}
}

// LoadEnglishMultitokenSpeller loads official resource files (Java file order).
// Empty paths are skipped. Returns a loaded speller or an empty one if no files open.
func LoadEnglishMultitokenSpeller(multiwordsPath, spellingGlobalPath string) (*EnglishMultitokenSpeller, error) {
	s := NewEnglishMultitokenSpeller()
	for _, p := range []string{multiwordsPath, spellingGlobalPath} {
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
