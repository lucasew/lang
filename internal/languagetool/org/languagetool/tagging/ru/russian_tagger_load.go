package ru

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

var (
	ruTaggerOnce  sync.Once
	ruPOSDictPath string
)

// DefaultRussianTagger is the process singleton (Java RussianTagger.INSTANCE).
// Backed by MapWordTagger until EnsureDefaultRussianTagger loads russian.dict.
var DefaultRussianTagger = NewRussianTagger(tagging.MapWordTagger{})

// DiscoverRussianPOSDict finds russian.dict (Java resource /ru/russian.dict).
// Order: LANG_RUSSIAN_DICT, walk-up inspiration module path.
func DiscoverRussianPOSDict() string {
	if p := os.Getenv("LANG_RUSSIAN_DICT"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	relPaths := []string{
		filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "ru",
			"src", "main", "resources", "org", "languagetool", "resource", "ru", "russian.dict"),
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

// EnsureDefaultRussianTagger loads russian.dict into DefaultRussianTagger.
// Idempotent; no-op if dict missing (fail closed).
func EnsureDefaultRussianTagger() {
	ruTaggerOnce.Do(func() {
		p := DiscoverRussianPOSDict()
		if p == "" {
			return
		}
		ruPOSDictPath = p
		mt := tagging.OpenMorfologikTagger(p)
		if mt == nil {
			return
		}
		// Java BaseTagger.initWordTagger: CombiningTagger when added.txt present
		wt := tagging.WordTagger(mt)
		if manual := loadRUManualTagger(p); manual != nil {
			removal := loadRUManualRemoval(p)
			wt = tagging.NewCombiningTaggerWithRemoval(wt, manual, removal, false)
		}
		DefaultRussianTagger = NewRussianTagger(wt)
	})
}

// RussianPOSDictPath returns the resolved russian.dict path after Ensure (may be empty).
func RussianPOSDictPath() string {
	EnsureDefaultRussianTagger()
	return ruPOSDictPath
}

func loadRUManualTagger(dictPath string) tagging.WordTagger {
	if wt := languagetool.LoadManualTaggerBesideDict(dictPath, []string{"added.txt", "added_custom.txt"}); wt != nil {
		return wt
	}
	if p := walkUpRUResourceDir(); p != "" {
		return languagetool.LoadManualTaggerFromDirs([]string{p}, []string{"added.txt", "added_custom.txt"})
	}
	return nil
}

func loadRUManualRemoval(dictPath string) tagging.WordTagger {
	if wt := languagetool.LoadManualTaggerBesideDict(dictPath, []string{"removed.txt", "removed_custom.txt"}); wt != nil {
		return wt
	}
	if p := walkUpRUResourceDir(); p != "" {
		return languagetool.LoadManualTaggerFromDirs([]string{p}, []string{"removed.txt", "removed_custom.txt"})
	}
	return nil
}

func walkUpRUResourceDir() string {
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "ru",
		"src", "main", "resources", "org", "languagetool", "resource", "ru")
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
