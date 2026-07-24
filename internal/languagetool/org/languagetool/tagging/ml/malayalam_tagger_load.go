package ml

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

var (
	mlTaggerOnce  sync.Once
	mlPOSDictPath string
)

// DefaultMalayalamTagger is the process singleton (Java new MalayalamTagger()).
// Backed by MapWordTagger until EnsureDefaultMalayalamTagger loads malayalam.dict.
var DefaultMalayalamTagger = NewMalayalamTagger(tagging.MapWordTagger{})

// DiscoverMalayalamPOSDict finds malayalam.dict (Java resource /ml/malayalam.dict).
// Order: LANG_MALAYALAM_DICT, walk-up inspiration module path.
func DiscoverMalayalamPOSDict() string {
	if p := os.Getenv("LANG_MALAYALAM_DICT"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	relPaths := []string{
		filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "ml",
			"src", "main", "resources", "org", "languagetool", "resource", "ml", "malayalam.dict"),
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

// EnsureDefaultMalayalamTagger loads malayalam.dict into DefaultMalayalamTagger.
// Idempotent; no-op if dict missing (fail closed).
func EnsureDefaultMalayalamTagger() {
	mlTaggerOnce.Do(func() {
		p := DiscoverMalayalamPOSDict()
		if p == "" {
			return
		}
		mlPOSDictPath = p
		mt := tagging.OpenMorfologikTagger(p)
		if mt == nil {
			return
		}
		// Java BaseTagger.initWordTagger: CombiningTagger when added.txt present
		wt := tagging.WordTagger(mt)
		if manual := loadMLManualTagger(p); manual != nil {
			removal := loadMLManualRemoval(p)
			wt = tagging.NewCombiningTaggerWithRemoval(wt, manual, removal, false)
		}
		DefaultMalayalamTagger = NewMalayalamTagger(wt)
	})
}

// MalayalamPOSDictPath returns the resolved malayalam.dict path after Ensure (may be empty).
func MalayalamPOSDictPath() string {
	EnsureDefaultMalayalamTagger()
	return mlPOSDictPath
}

func loadMLManualTagger(dictPath string) tagging.WordTagger {
	if wt := languagetool.LoadManualTaggerBesideDict(dictPath, []string{"added.txt", "added_custom.txt"}); wt != nil {
		return wt
	}
	if p := walkUpMLResourceDir(); p != "" {
		return languagetool.LoadManualTaggerFromDirs([]string{p}, []string{"added.txt", "added_custom.txt"})
	}
	return nil
}

func loadMLManualRemoval(dictPath string) tagging.WordTagger {
	if wt := languagetool.LoadManualTaggerBesideDict(dictPath, []string{"removed.txt", "removed_custom.txt"}); wt != nil {
		return wt
	}
	if p := walkUpMLResourceDir(); p != "" {
		return languagetool.LoadManualTaggerFromDirs([]string{p}, []string{"removed.txt", "removed_custom.txt"})
	}
	return nil
}

func walkUpMLResourceDir() string {
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "ml",
		"src", "main", "resources", "org", "languagetool", "resource", "ml")
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
