package el

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

var (
	elTaggerOnce  sync.Once
	elPOSDictPath string
)

// DefaultGreekTagger is the process singleton (Java new GreekTagger()).
// Backed by MapWordTagger until EnsureDefaultGreekTagger loads greek.dict.
var DefaultGreekTagger = NewGreekTagger(tagging.MapWordTagger{})

// DiscoverGreekPOSDict finds greek.dict (Java resource /el/greek.dict).
// Order: LANG_GREEK_DICT / LANG_EL_DICT, walk-up inspiration module path.
func DiscoverGreekPOSDict() string {
	for _, env := range []string{"LANG_GREEK_DICT", "LANG_EL_DICT"} {
		if p := os.Getenv(env); p != "" {
			if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
				return p
			}
		}
	}
	relPaths := []string{
		filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "el",
			"src", "main", "resources", "org", "languagetool", "resource", "el", "greek.dict"),
	}
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for i := 0; i < 14; i++ {
		for _, rel := range relPaths {
			p := filepath.Join(dir, rel)
			if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
				return p
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// EnsureDefaultGreekTagger loads greek.dict into DefaultGreekTagger.
// Idempotent; no-op if dict missing (fail closed).
// Java also wires GreekAnalyzer.additionalTags — not invent-ported here.
func EnsureDefaultGreekTagger() {
	elTaggerOnce.Do(func() {
		p := DiscoverGreekPOSDict()
		if p == "" {
			return
		}
		elPOSDictPath = p
		mt := tagging.OpenMorfologikTagger(p)
		if mt == nil {
			return
		}
		// Java BaseTagger.initWordTagger: CombiningTagger when added.txt present
		wt := tagging.WordTagger(mt)
		if manual := loadELManualTagger(p); manual != nil {
			removal := loadELManualRemoval(p)
			wt = tagging.NewCombiningTaggerWithRemoval(wt, manual, removal, false)
		}
		DefaultGreekTagger = NewGreekTagger(wt)
	})
}

// GreekPOSDictPath returns the resolved greek.dict path after Ensure (may be empty).
func GreekPOSDictPath() string {
	EnsureDefaultGreekTagger()
	return elPOSDictPath
}

func loadELManualTagger(dictPath string) tagging.WordTagger {
	if wt := languagetool.LoadManualTaggerBesideDict(dictPath, []string{"added.txt", "added_custom.txt"}); wt != nil {
		return wt
	}
	if p := walkUpELResourceDir(); p != "" {
		return languagetool.LoadManualTaggerFromDirs([]string{p}, []string{"added.txt", "added_custom.txt"})
	}
	return nil
}

func loadELManualRemoval(dictPath string) tagging.WordTagger {
	if wt := languagetool.LoadManualTaggerBesideDict(dictPath, []string{"removed.txt", "removed_custom.txt"}); wt != nil {
		return wt
	}
	if p := walkUpELResourceDir(); p != "" {
		return languagetool.LoadManualTaggerFromDirs([]string{p}, []string{"removed.txt", "removed_custom.txt"})
	}
	return nil
}

func walkUpELResourceDir() string {
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "el",
		"src", "main", "resources", "org", "languagetool", "resource", "el")
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for i := 0; i < 14; i++ {
		p := filepath.Join(dir, rel)
		if st, err := os.Stat(p); err == nil && st.IsDir() {
			return p
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}
