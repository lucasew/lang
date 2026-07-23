package km

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

var (
	kmTaggerOnce  sync.Once
	kmPOSDictPath string
)

// DefaultKhmerTagger is the process singleton (Java new KhmerTagger()).
// Backed by MapWordTagger until EnsureDefaultKhmerTagger loads khmer.dict.
var DefaultKhmerTagger = NewKhmerTagger(tagging.MapWordTagger{})

// DiscoverKhmerPOSDict finds khmer.dict (Java resource /km/khmer.dict).
// Order: LANG_KHMER_DICT, walk-up inspiration module path.
func DiscoverKhmerPOSDict() string {
	if p := os.Getenv("LANG_KHMER_DICT"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	relPaths := []string{
		filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "km",
			"src", "main", "resources", "org", "languagetool", "resource", "km", "khmer.dict"),
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

// EnsureDefaultKhmerTagger loads khmer.dict into DefaultKhmerTagger.
// Idempotent; no-op if dict missing (fail closed).
func EnsureDefaultKhmerTagger() {
	kmTaggerOnce.Do(func() {
		p := DiscoverKhmerPOSDict()
		if p == "" {
			return
		}
		kmPOSDictPath = p
		mt := tagging.OpenMorfologikTagger(p)
		if mt == nil {
			return
		}
		// Java BaseTagger.initWordTagger: CombiningTagger when added.txt present
		wt := tagging.WordTagger(mt)
		if manual := loadKMManualTagger(p); manual != nil {
			removal := loadKMManualRemoval(p)
			wt = tagging.NewCombiningTaggerWithRemoval(wt, manual, removal, false)
		}
		DefaultKhmerTagger = NewKhmerTagger(wt)
	})
}

// KhmerPOSDictPath returns the resolved khmer.dict path after Ensure (may be empty).
func KhmerPOSDictPath() string {
	EnsureDefaultKhmerTagger()
	return kmPOSDictPath
}

func loadKMManualTagger(dictPath string) tagging.WordTagger {
	if wt := languagetool.LoadManualTaggerBesideDict(dictPath, []string{"added.txt", "added_custom.txt"}); wt != nil {
		return wt
	}
	if p := walkUpKMResourceDir(); p != "" {
		return languagetool.LoadManualTaggerFromDirs([]string{p}, []string{"added.txt", "added_custom.txt"})
	}
	return nil
}

func loadKMManualRemoval(dictPath string) tagging.WordTagger {
	if wt := languagetool.LoadManualTaggerBesideDict(dictPath, []string{"removed.txt", "removed_custom.txt"}); wt != nil {
		return wt
	}
	if p := walkUpKMResourceDir(); p != "" {
		return languagetool.LoadManualTaggerFromDirs([]string{p}, []string{"removed.txt", "removed_custom.txt"})
	}
	return nil
}

func walkUpKMResourceDir() string {
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "km",
		"src", "main", "resources", "org", "languagetool", "resource", "km")
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
