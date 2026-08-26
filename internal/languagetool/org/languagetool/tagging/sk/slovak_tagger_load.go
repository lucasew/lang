package sk

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

var (
	skTaggerOnce  sync.Once
	skPOSDictPath string
)

// DefaultSlovakTagger is the process singleton (Java new SlovakTagger()).
// Backed by MapWordTagger until EnsureDefaultSlovakTagger loads slovak.dict.
var DefaultSlovakTagger = NewSlovakTagger(tagging.MapWordTagger{})

// DiscoverSlovakPOSDict finds slovak.dict (Java resource /sk/slovak.dict).
// Order: LANG_SLOVAK_DICT, walk-up inspiration module path.
func DiscoverSlovakPOSDict() string {
	if p := os.Getenv("LANG_SLOVAK_DICT"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	relPaths := []string{
		filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "sk",
			"src", "main", "resources", "org", "languagetool", "resource", "sk", "slovak.dict"),
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

// EnsureDefaultSlovakTagger loads slovak.dict into DefaultSlovakTagger.
// Idempotent; no-op if dict missing (fail closed).
func EnsureDefaultSlovakTagger() {
	skTaggerOnce.Do(func() {
		p := DiscoverSlovakPOSDict()
		if p == "" {
			return
		}
		skPOSDictPath = p
		mt := tagging.OpenMorfologikTagger(p)
		if mt == nil {
			return
		}
		// Java BaseTagger.initWordTagger: CombiningTagger when added.txt present
		wt := tagging.WordTagger(mt)
		if manual := loadSKManualTagger(p); manual != nil {
			removal := loadSKManualRemoval(p)
			wt = tagging.NewCombiningTaggerWithRemoval(wt, manual, removal, false)
		}
		DefaultSlovakTagger = NewSlovakTagger(wt)
	})
}

// SlovakPOSDictPath returns the resolved slovak.dict path after Ensure (may be empty).
func SlovakPOSDictPath() string {
	EnsureDefaultSlovakTagger()
	return skPOSDictPath
}

func loadSKManualTagger(dictPath string) tagging.WordTagger {
	if wt := languagetool.LoadManualTaggerBesideDict(dictPath, []string{"added.txt", "added_custom.txt"}); wt != nil {
		return wt
	}
	if p := walkUpSKResourceDir(); p != "" {
		return languagetool.LoadManualTaggerFromDirs([]string{p}, []string{"added.txt", "added_custom.txt"})
	}
	return nil
}

func loadSKManualRemoval(dictPath string) tagging.WordTagger {
	if wt := languagetool.LoadManualTaggerBesideDict(dictPath, []string{"removed.txt", "removed_custom.txt"}); wt != nil {
		return wt
	}
	if p := walkUpSKResourceDir(); p != "" {
		return languagetool.LoadManualTaggerFromDirs([]string{p}, []string{"removed.txt", "removed_custom.txt"})
	}
	return nil
}

func walkUpSKResourceDir() string {
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "sk",
		"src", "main", "resources", "org", "languagetool", "resource", "sk")
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
