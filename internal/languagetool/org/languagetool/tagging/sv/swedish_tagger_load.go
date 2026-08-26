package sv

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

var (
	svTaggerOnce  sync.Once
	svPOSDictPath string
)

// DefaultSwedishTagger is the process singleton (Java new SwedishTagger()).
// Backed by MapWordTagger until EnsureDefaultSwedishTagger loads swedish.dict.
var DefaultSwedishTagger = NewSwedishTagger(tagging.MapWordTagger{})

// DiscoverSwedishPOSDict finds swedish.dict (Java resource /sv/swedish.dict).
// Order: LANG_SWEDISH_DICT, walk-up inspiration module path.
func DiscoverSwedishPOSDict() string {
	if p := os.Getenv("LANG_SWEDISH_DICT"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	relPaths := []string{
		filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "sv",
			"src", "main", "resources", "org", "languagetool", "resource", "sv", "swedish.dict"),
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

// EnsureDefaultSwedishTagger loads swedish.dict into DefaultSwedishTagger.
// Idempotent; no-op if dict missing (fail closed).
func EnsureDefaultSwedishTagger() {
	svTaggerOnce.Do(func() {
		p := DiscoverSwedishPOSDict()
		if p == "" {
			return
		}
		svPOSDictPath = p
		mt := tagging.OpenMorfologikTagger(p)
		if mt == nil {
			return
		}
		// Java BaseTagger.initWordTagger: CombiningTagger when added.txt present
		wt := tagging.WordTagger(mt)
		if manual := loadSVManualTagger(p); manual != nil {
			removal := loadSVManualRemoval(p)
			wt = tagging.NewCombiningTaggerWithRemoval(wt, manual, removal, false)
		}
		DefaultSwedishTagger = NewSwedishTagger(wt)
	})
}

// SwedishPOSDictPath returns the resolved swedish.dict path after Ensure (may be empty).
func SwedishPOSDictPath() string {
	EnsureDefaultSwedishTagger()
	return svPOSDictPath
}

func loadSVManualTagger(dictPath string) tagging.WordTagger {
	if wt := languagetool.LoadManualTaggerBesideDict(dictPath, []string{"added.txt", "added_custom.txt"}); wt != nil {
		return wt
	}
	if p := walkUpSVResourceDir(); p != "" {
		return languagetool.LoadManualTaggerFromDirs([]string{p}, []string{"added.txt", "added_custom.txt"})
	}
	return nil
}

func loadSVManualRemoval(dictPath string) tagging.WordTagger {
	if wt := languagetool.LoadManualTaggerBesideDict(dictPath, []string{"removed.txt", "removed_custom.txt"}); wt != nil {
		return wt
	}
	if p := walkUpSVResourceDir(); p != "" {
		return languagetool.LoadManualTaggerFromDirs([]string{p}, []string{"removed.txt", "removed_custom.txt"})
	}
	return nil
}

func walkUpSVResourceDir() string {
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "sv",
		"src", "main", "resources", "org", "languagetool", "resource", "sv")
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
