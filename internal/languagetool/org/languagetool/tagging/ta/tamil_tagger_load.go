package ta

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

var (
	taTaggerOnce  sync.Once
	taPOSDictPath string
)

// DefaultTamilTagger is the process singleton (Java new TamilTagger()).
// Backed by MapWordTagger until EnsureDefaultTamilTagger loads tamil.dict.
var DefaultTamilTagger = NewTamilTagger(tagging.MapWordTagger{})

// DiscoverTamilPOSDict finds tamil.dict (Java resource /ta/tamil.dict).
// Order: LANG_TAMIL_DICT, walk-up inspiration module path.
func DiscoverTamilPOSDict() string {
	if p := os.Getenv("LANG_TAMIL_DICT"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	relPaths := []string{
		filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "ta",
			"src", "main", "resources", "org", "languagetool", "resource", "ta", "tamil.dict"),
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

// EnsureDefaultTamilTagger loads tamil.dict into DefaultTamilTagger.
// Idempotent; no-op if dict missing (fail closed).
func EnsureDefaultTamilTagger() {
	taTaggerOnce.Do(func() {
		p := DiscoverTamilPOSDict()
		if p == "" {
			return
		}
		taPOSDictPath = p
		mt := tagging.OpenMorfologikTagger(p)
		if mt == nil {
			return
		}
		// Java BaseTagger.initWordTagger: CombiningTagger when added.txt present
		wt := tagging.WordTagger(mt)
		if manual := loadTAManualTagger(p); manual != nil {
			removal := loadTAManualRemoval(p)
			wt = tagging.NewCombiningTaggerWithRemoval(wt, manual, removal, false)
		}
		DefaultTamilTagger = NewTamilTagger(wt)
	})
}

// TamilPOSDictPath returns the resolved tamil.dict path after Ensure (may be empty).
func TamilPOSDictPath() string {
	EnsureDefaultTamilTagger()
	return taPOSDictPath
}

func loadTAManualTagger(dictPath string) tagging.WordTagger {
	if wt := languagetool.LoadManualTaggerBesideDict(dictPath, []string{"added.txt", "added_custom.txt"}); wt != nil {
		return wt
	}
	if p := walkUpTAResourceDir(); p != "" {
		return languagetool.LoadManualTaggerFromDirs([]string{p}, []string{"added.txt", "added_custom.txt"})
	}
	return nil
}

func loadTAManualRemoval(dictPath string) tagging.WordTagger {
	if wt := languagetool.LoadManualTaggerBesideDict(dictPath, []string{"removed.txt", "removed_custom.txt"}); wt != nil {
		return wt
	}
	if p := walkUpTAResourceDir(); p != "" {
		return languagetool.LoadManualTaggerFromDirs([]string{p}, []string{"removed.txt", "removed_custom.txt"})
	}
	return nil
}

func walkUpTAResourceDir() string {
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "ta",
		"src", "main", "resources", "org", "languagetool", "resource", "ta")
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
