package ro

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

var (
	roTaggerOnce  sync.Once
	roPOSDictPath string
)

// DefaultRomanianTagger is the process singleton (Java new RomanianTagger()).
// Backed by MapWordTagger until EnsureDefaultRomanianTagger loads romanian.dict.
var DefaultRomanianTagger = NewRomanianTagger(tagging.MapWordTagger{})

// DiscoverRomanianPOSDict finds romanian.dict (Java resource /ro/romanian.dict).
// Order: LANG_ROMANIAN_DICT, walk-up inspiration module path.
func DiscoverRomanianPOSDict() string {
	if p := os.Getenv("LANG_ROMANIAN_DICT"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	return walkUpROFile("romanian.dict")
}

// DiscoverRomanianDiacriticsDict finds test_diacritics.dict
// (Java resource /ro/test_diacritics.dict used by RomanianTaggerDiacriticsTest).
func DiscoverRomanianDiacriticsDict() string {
	if p := os.Getenv("LANG_ROMANIAN_DIACRITICS_DICT"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	return walkUpROFile("test_diacritics.dict")
}

// EnsureDefaultRomanianTagger loads romanian.dict into DefaultRomanianTagger.
// Idempotent; no-op if dict missing (fail closed).
func EnsureDefaultRomanianTagger() {
	roTaggerOnce.Do(func() {
		p := DiscoverRomanianPOSDict()
		if p == "" {
			return
		}
		roPOSDictPath = p
		if tg := loadRomanianTaggerAt(p, RomanianDictPath); tg != nil {
			DefaultRomanianTagger = tg
		}
	})
}

// RomanianPOSDictPath returns the resolved romanian.dict path after Ensure (may be empty).
func RomanianPOSDictPath() string {
	EnsureDefaultRomanianTagger()
	return roPOSDictPath
}

// OpenRomanianTaggerFromFilesystem loads a RomanianTagger from a filesystem .dict path
// with the given Java resource dictionary path (e.g. RomanianTestDiacriticsDictPath).
// Manual added/removed still come from the ro resource dir (Java locale-based manuals).
// Returns nil if the binary dict cannot be opened.
func OpenRomanianTaggerFromFilesystem(fsDictPath, resourceDictPath string) *RomanianTagger {
	return loadRomanianTaggerAt(fsDictPath, resourceDictPath)
}

func loadRomanianTaggerAt(fsDictPath, resourceDictPath string) *RomanianTagger {
	if fsDictPath == "" {
		return nil
	}
	mt := tagging.OpenMorfologikTagger(fsDictPath)
	if mt == nil {
		return nil
	}
	// Java BaseTagger.initWordTagger: CombiningTagger when added.txt present
	wt := tagging.WordTagger(mt)
	if manual := loadROManualTagger(fsDictPath); manual != nil {
		removal := loadROManualRemoval(fsDictPath)
		wt = tagging.NewCombiningTaggerWithRemoval(wt, manual, removal, false)
	}
	return NewRomanianTaggerWithPath(wt, resourceDictPath)
}

func loadROManualTagger(dictPath string) tagging.WordTagger {
	if wt := languagetool.LoadManualTaggerBesideDict(dictPath, []string{"added.txt", "added_custom.txt"}); wt != nil {
		return wt
	}
	if p := walkUpROResourceDir(); p != "" {
		return languagetool.LoadManualTaggerFromDirs([]string{p}, []string{"added.txt", "added_custom.txt"})
	}
	return nil
}

func loadROManualRemoval(dictPath string) tagging.WordTagger {
	if wt := languagetool.LoadManualTaggerBesideDict(dictPath, []string{"removed.txt", "removed_custom.txt"}); wt != nil {
		return wt
	}
	if p := walkUpROResourceDir(); p != "" {
		return languagetool.LoadManualTaggerFromDirs([]string{p}, []string{"removed.txt", "removed_custom.txt"})
	}
	return nil
}

func walkUpROFile(name string) string {
	relPaths := []string{
		filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "ro",
			"src", "main", "resources", "org", "languagetool", "resource", "ro", name),
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

func walkUpROResourceDir() string {
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "ro",
		"src", "main", "resources", "org", "languagetool", "resource", "ro")
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
