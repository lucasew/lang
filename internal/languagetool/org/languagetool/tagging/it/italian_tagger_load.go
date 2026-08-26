package it

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

var (
	itTaggerOnce  sync.Once
	itPOSDictPath string
)

// DefaultItalianTagger is the process singleton (Java new ItalianTagger()).
// Backed by MapWordTagger until EnsureDefaultItalianTagger loads italian.dict.
var DefaultItalianTagger = NewItalianTagger(tagging.MapWordTagger{})

// DiscoverItalianPOSDict finds italian.dict (Java resource /it/italian.dict).
// Order: LANG_ITALIAN_DICT, walk-up inspiration module path.
func DiscoverItalianPOSDict() string {
	if p := os.Getenv("LANG_ITALIAN_DICT"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	relPaths := []string{
		filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "it",
			"src", "main", "resources", "org", "languagetool", "resource", "it", "italian.dict"),
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

// EnsureDefaultItalianTagger loads italian.dict into DefaultItalianTagger.
// Idempotent; no-op if dict missing (fail closed).
func EnsureDefaultItalianTagger() {
	itTaggerOnce.Do(func() {
		p := DiscoverItalianPOSDict()
		if p == "" {
			return
		}
		itPOSDictPath = p
		mt := tagging.OpenMorfologikTagger(p)
		if mt == nil {
			return
		}
		// Java BaseTagger.initWordTagger: CombiningTagger when added.txt present
		wt := tagging.WordTagger(mt)
		if manual := loadITManualTagger(p); manual != nil {
			removal := loadITManualRemoval(p)
			wt = tagging.NewCombiningTaggerWithRemoval(wt, manual, removal, false)
		}
		DefaultItalianTagger = NewItalianTagger(wt)
	})
}

// ItalianPOSDictPath returns the resolved italian.dict path after Ensure (may be empty).
func ItalianPOSDictPath() string {
	EnsureDefaultItalianTagger()
	return itPOSDictPath
}

func loadITManualTagger(dictPath string) tagging.WordTagger {
	if wt := languagetool.LoadManualTaggerBesideDict(dictPath, []string{"added.txt", "added_custom.txt"}); wt != nil {
		return wt
	}
	if p := walkUpITResourceDir(); p != "" {
		return languagetool.LoadManualTaggerFromDirs([]string{p}, []string{"added.txt", "added_custom.txt"})
	}
	return nil
}

func loadITManualRemoval(dictPath string) tagging.WordTagger {
	if wt := languagetool.LoadManualTaggerBesideDict(dictPath, []string{"removed.txt", "removed_custom.txt"}); wt != nil {
		return wt
	}
	if p := walkUpITResourceDir(); p != "" {
		return languagetool.LoadManualTaggerFromDirs([]string{p}, []string{"removed.txt", "removed_custom.txt"})
	}
	return nil
}

func walkUpITResourceDir() string {
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "it",
		"src", "main", "resources", "org", "languagetool", "resource", "it")
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
