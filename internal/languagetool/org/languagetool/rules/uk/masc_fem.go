package uk

import (
	"embed"
	"strings"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

//go:embed data/masc_fem.txt
var mascFemFS embed.FS

var (
	mascFemOnce sync.Once
	mascFemSet  map[string]struct{}
)

// loadMascFemSet ports MASC_FEM_SET = extendSet(loadSet("/uk/masc_fem.txt"), "екс-").
func loadMascFemSet() map[string]struct{} {
	mascFemOnce.Do(func() {
		f, err := mascFemFS.Open("data/masc_fem.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()
		base, err := LoadSet(f)
		if err != nil {
			panic(err)
		}
		// Java extendSet: also add "екс-" + each line
		out := make(map[string]struct{}, len(base)*2)
		for k := range base {
			out[k] = struct{}{}
			out["екс-"+k] = struct{}{}
		}
		mascFemSet = out
	})
	return mascFemSet
}

// IsInMascFemSet ports isInMascFemSet (normalize curly apostrophe to hyphen).
func IsInMascFemSet(lemma string) bool {
	if lemma == "" {
		return false
	}
	lemma = strings.ReplaceAll(lemma, "\u2018", "-")
	_, ok := loadMascFemSet()[lemma]
	return ok
}

// HasMascFemLemma ports hasMascFemLemma(List<AnalyzedToken>).
// Male anim nominative professions that may refer to women (fem verb agreement OK).
func HasMascFemLemma(tok *languagetool.AnalyzedTokenReadings) bool {
	if tok == nil {
		return false
	}
	surface := tok.GetToken()
	if strings.HasSuffix(surface, "олог") || strings.HasSuffix(surface, "знавець") {
		return true
	}
	for _, r := range tok.GetReadings() {
		if r == nil || r.GetPOSTag() == nil || r.GetLemma() == nil {
			continue
		}
		if !strings.Contains(*r.GetPOSTag(), "noun:anim:m:v_naz") {
			continue
		}
		lemma := *r.GetLemma()
		if IsInMascFemSet(lemma) {
			return true
		}
		if i := strings.Index(lemma, "-"); i > 0 {
			if IsInMascFemSet(lemma[:i]) {
				return true
			}
		}
	}
	return false
}
