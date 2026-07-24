package ar

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

var (
	arTaggerOnce  sync.Once
	arPOSDictPath string
)

// DefaultArabicTagger is the process singleton (Java new ArabicTagger()).
// Backed by MapWordTagger until EnsureDefaultArabicTagger loads arabic.dict.
var DefaultArabicTagger = NewArabicTagger(tagging.MapWordTagger{})

// DiscoverArabicPOSDict finds arabic.dict (Java resource /ar/arabic.dict).
// Order: LANG_ARABIC_DICT, walk-up inspiration module path.
func DiscoverArabicPOSDict() string {
	if p := os.Getenv("LANG_ARABIC_DICT"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	relPaths := []string{
		filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "ar",
			"src", "main", "resources", "org", "languagetool", "resource", "ar", "arabic.dict"),
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

// EnsureDefaultArabicTagger loads arabic.dict into DefaultArabicTagger.
// Idempotent; no-op if dict missing (fail closed).
// Java: BaseTagger initWordTagger with overwriteWithManualTagger() → false (default).
// additionalTags uses DictionaryLookup(getDictionary()) — binary dict only.
func EnsureDefaultArabicTagger() {
	arTaggerOnce.Do(func() {
		p := DiscoverArabicPOSDict()
		if p == "" {
			return
		}
		arPOSDictPath = p
		mt := tagging.OpenMorfologikTagger(p)
		if mt == nil {
			return
		}
		// Java BaseTagger.initWordTagger: CombiningTagger when added.txt present
		wt := tagging.WordTagger(mt)
		if manual := loadARManualTagger(p); manual != nil {
			removal := loadARManualRemoval(p)
			wt = tagging.NewCombiningTaggerWithRemoval(wt, manual, removal, false)
		}
		// dictLookup = DictionaryLookup(getDictionary()) for additionalTags
		DefaultArabicTagger = NewArabicTaggerWithDictLookup(wt, mt)
	})
}

// ArabicPOSDictPath returns the resolved arabic.dict path after Ensure (may be empty).
func ArabicPOSDictPath() string {
	EnsureDefaultArabicTagger()
	return arPOSDictPath
}

func loadARManualTagger(dictPath string) tagging.WordTagger {
	if wt := languagetool.LoadManualTaggerBesideDict(dictPath, []string{"added.txt", "added_custom.txt"}); wt != nil {
		return wt
	}
	if p := walkUpARResourceDir(); p != "" {
		return languagetool.LoadManualTaggerFromDirs([]string{p}, []string{"added.txt", "added_custom.txt"})
	}
	return nil
}

func loadARManualRemoval(dictPath string) tagging.WordTagger {
	if wt := languagetool.LoadManualTaggerBesideDict(dictPath, []string{"removed.txt", "removed_custom.txt"}); wt != nil {
		return wt
	}
	if p := walkUpARResourceDir(); p != "" {
		return languagetool.LoadManualTaggerFromDirs([]string{p}, []string{"removed.txt", "removed_custom.txt"})
	}
	return nil
}

func walkUpARResourceDir() string {
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "ar",
		"src", "main", "resources", "org", "languagetool", "resource", "ar")
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
