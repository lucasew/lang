package da

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

var (
	daTaggerOnce  sync.Once
	daPOSDictPath string
)

// DefaultDanishTagger is the process singleton (Java new DanishTagger()).
// Backed by MapWordTagger until EnsureDefaultDanishTagger loads danish.dict.
var DefaultDanishTagger = NewDanishTagger(tagging.MapWordTagger{})

// DiscoverDanishPOSDict finds danish.dict (Java resource /da/danish.dict).
// Order: LANG_DANISH_DICT, walk-up inspiration module path.
func DiscoverDanishPOSDict() string {
	if p := os.Getenv("LANG_DANISH_DICT"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	relPaths := []string{
		filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "da",
			"src", "main", "resources", "org", "languagetool", "resource", "da", "danish.dict"),
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

// EnsureDefaultDanishTagger loads danish.dict into DefaultDanishTagger.
// Idempotent; no-op if dict missing (fail closed).
func EnsureDefaultDanishTagger() {
	daTaggerOnce.Do(func() {
		p := DiscoverDanishPOSDict()
		if p == "" {
			return
		}
		daPOSDictPath = p
		mt := tagging.OpenMorfologikTagger(p)
		if mt == nil {
			return
		}
		// Java BaseTagger.initWordTagger: CombiningTagger when added.txt present
		wt := tagging.WordTagger(mt)
		if manual := loadDAManualTagger(p); manual != nil {
			removal := loadDAManualRemoval(p)
			wt = tagging.NewCombiningTaggerWithRemoval(wt, manual, removal, false)
		}
		DefaultDanishTagger = NewDanishTagger(wt)
	})
}

// DanishPOSDictPath returns the resolved danish.dict path after Ensure (may be empty).
func DanishPOSDictPath() string {
	EnsureDefaultDanishTagger()
	return daPOSDictPath
}

func loadDAManualTagger(dictPath string) tagging.WordTagger {
	if wt := languagetool.LoadManualTaggerBesideDict(dictPath, []string{"added.txt", "added_custom.txt"}); wt != nil {
		return wt
	}
	if p := walkUpDAResourceDir(); p != "" {
		return languagetool.LoadManualTaggerFromDirs([]string{p}, []string{"added.txt", "added_custom.txt"})
	}
	return nil
}

func loadDAManualRemoval(dictPath string) tagging.WordTagger {
	if wt := languagetool.LoadManualTaggerBesideDict(dictPath, []string{"removed.txt", "removed_custom.txt"}); wt != nil {
		return wt
	}
	if p := walkUpDAResourceDir(); p != "" {
		return languagetool.LoadManualTaggerFromDirs([]string{p}, []string{"removed.txt", "removed_custom.txt"})
	}
	return nil
}

func walkUpDAResourceDir() string {
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "da",
		"src", "main", "resources", "org", "languagetool", "resource", "da")
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
