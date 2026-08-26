package pl

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

var (
	plTaggerOnce  sync.Once
	plPOSDictPath string
	// DefaultPolishTagger is the process-wide PolishTagger backed by polish.dict
	// (Java: new Polish().getTagger()). Nil if dict is missing.
	DefaultPolishTagger *PolishTagger
)

// DiscoverPolishPOSDict finds polish.dict (Java resource /pl/polish.dict).
// Order: LANG_POLISH_DICT, walk-up inspiration module path.
func DiscoverPolishPOSDict() string {
	if p := os.Getenv("LANG_POLISH_DICT"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	relPaths := []string{
		filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "pl",
			"src", "main", "resources", "org", "languagetool", "resource", "pl", "polish.dict"),
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

// EnsureDefaultPolishTagger loads polish.dict into DefaultPolishTagger.
// Idempotent; no-op if dict missing (fail closed).
func EnsureDefaultPolishTagger() {
	plTaggerOnce.Do(func() {
		p := DiscoverPolishPOSDict()
		if p == "" {
			return
		}
		plPOSDictPath = p
		mt := tagging.OpenMorfologikTagger(p)
		if mt == nil {
			return
		}
		// Java BaseTagger.initWordTagger: CombiningTagger when added.txt present
		wt := tagging.WordTagger(mt)
		if manual := loadPLManualTagger(p); manual != nil {
			removal := loadPLManualRemoval(p)
			wt = tagging.NewCombiningTaggerWithRemoval(wt, manual, removal, false)
		}
		DefaultPolishTagger = NewPolishTagger(wt)
	})
}

// PolishPOSDictPath returns the resolved polish.dict path after Ensure (may be empty).
func PolishPOSDictPath() string {
	EnsureDefaultPolishTagger()
	return plPOSDictPath
}

func loadPLManualTagger(dictPath string) tagging.WordTagger {
	if wt := languagetool.LoadManualTaggerBesideDict(dictPath, []string{"added.txt", "added_custom.txt"}); wt != nil {
		return wt
	}
	if p := walkUpPLResourceDir(); p != "" {
		return languagetool.LoadManualTaggerFromDirs([]string{p}, []string{"added.txt", "added_custom.txt"})
	}
	return nil
}

func loadPLManualRemoval(dictPath string) tagging.WordTagger {
	if wt := languagetool.LoadManualTaggerBesideDict(dictPath, []string{"removed.txt", "removed_custom.txt"}); wt != nil {
		return wt
	}
	if p := walkUpPLResourceDir(); p != "" {
		return languagetool.LoadManualTaggerFromDirs([]string{p}, []string{"removed.txt", "removed_custom.txt"})
	}
	return nil
}

func walkUpPLResourceDir() string {
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "pl",
		"src", "main", "resources", "org", "languagetool", "resource", "pl")
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
