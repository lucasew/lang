package br

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

var (
	brTaggerOnce  sync.Once
	brPOSDictPath string
)

// DefaultBretonTagger is the process singleton (Java new BretonTagger()).
// Backed by MapWordTagger until EnsureDefaultBretonTagger loads breton.dict.
var DefaultBretonTagger = NewBretonTagger(tagging.MapWordTagger{})

// DiscoverBretonPOSDict finds breton.dict (Java resource /br/breton.dict).
// Order: LANG_BRETON_DICT, walk-up inspiration module path.
func DiscoverBretonPOSDict() string {
	if p := os.Getenv("LANG_BRETON_DICT"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	relPaths := []string{
		filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "br",
			"src", "main", "resources", "org", "languagetool", "resource", "br", "breton.dict"),
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

// EnsureDefaultBretonTagger loads breton.dict into DefaultBretonTagger.
// Idempotent; no-op if dict missing (fail closed).
func EnsureDefaultBretonTagger() {
	brTaggerOnce.Do(func() {
		p := DiscoverBretonPOSDict()
		if p == "" {
			return
		}
		brPOSDictPath = p
		mt := tagging.OpenMorfologikTagger(p)
		if mt == nil {
			return
		}
		// Java BaseTagger.initWordTagger: CombiningTagger when added.txt present
		wt := tagging.WordTagger(mt)
		if manual := loadBRManualTagger(p); manual != nil {
			removal := loadBRManualRemoval(p)
			wt = tagging.NewCombiningTaggerWithRemoval(wt, manual, removal, false)
		}
		DefaultBretonTagger = NewBretonTagger(wt)
	})
}

// BretonPOSDictPath returns the resolved breton.dict path after Ensure (may be empty).
func BretonPOSDictPath() string {
	EnsureDefaultBretonTagger()
	return brPOSDictPath
}

func loadBRManualTagger(dictPath string) tagging.WordTagger {
	if wt := languagetool.LoadManualTaggerBesideDict(dictPath, []string{"added.txt", "added_custom.txt"}); wt != nil {
		return wt
	}
	if p := walkUpBRResourceDir(); p != "" {
		return languagetool.LoadManualTaggerFromDirs([]string{p}, []string{"added.txt", "added_custom.txt"})
	}
	return nil
}

func loadBRManualRemoval(dictPath string) tagging.WordTagger {
	if wt := languagetool.LoadManualTaggerBesideDict(dictPath, []string{"removed.txt", "removed_custom.txt"}); wt != nil {
		return wt
	}
	if p := walkUpBRResourceDir(); p != "" {
		return languagetool.LoadManualTaggerFromDirs([]string{p}, []string{"removed.txt", "removed_custom.txt"})
	}
	return nil
}

func walkUpBRResourceDir() string {
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "br",
		"src", "main", "resources", "org", "languagetool", "resource", "br")
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
