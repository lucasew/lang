package ga

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

var (
	gaTaggerOnce  sync.Once
	gaPOSDictPath string
)

// DefaultIrishTagger is the process singleton (Java new IrishTagger()).
// Backed by MapWordTagger until EnsureDefaultIrishTagger loads irish.dict.
var DefaultIrishTagger = NewIrishTagger(tagging.MapWordTagger{})

// DiscoverIrishPOSDict finds irish.dict (Java resource /ga/irish.dict).
// Order: LANG_IRISH_DICT, LANG_GA_DICT, /tmp Maven extract, third_party, inspiration walk-up.
func DiscoverIrishPOSDict() string {
	for _, env := range []string{"LANG_IRISH_DICT", "LANG_GA_DICT"} {
		if p := os.Getenv(env); p != "" {
			if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
				return p
			}
		}
	}
	// Pre-extracted Maven languagetool-ga-dicts (official irish.dict bytes).
	absCandidates := []string{
		"/tmp/ga-dict-fetch/org/languagetool/resource/ga/irish.dict",
	}
	for _, p := range absCandidates {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	relPaths := []string{
		filepath.Join("third_party", "irish-pos-dict", "org", "languagetool", "resource", "ga", "irish.dict"),
		filepath.Join("third_party", "ga-pos-dict", "org", "languagetool", "resource", "ga", "irish.dict"),
		filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "ga",
			"src", "main", "resources", "org", "languagetool", "resource", "ga", "irish.dict"),
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

// EnsureDefaultIrishTagger loads irish.dict into DefaultIrishTagger.
// Idempotent; no-op if dict missing (fail closed).
// Java BaseTagger: overwriteWithManualTagger() → false → CombiningTagger(overwrite=false).
func EnsureDefaultIrishTagger() {
	gaTaggerOnce.Do(func() {
		p := DiscoverIrishPOSDict()
		if p == "" {
			return
		}
		gaPOSDictPath = p
		mt := tagging.OpenMorfologikTagger(p)
		if mt == nil {
			return
		}
		// Java BaseTagger.initWordTagger: CombiningTagger when added.txt present
		wt := tagging.WordTagger(mt)
		if manual := loadGAManualTagger(p); manual != nil {
			removal := loadGAManualRemoval(p)
			wt = tagging.NewCombiningTaggerWithRemoval(wt, manual, removal, false)
		}
		DefaultIrishTagger = NewIrishTagger(wt)
	})
}

// IrishPOSDictPath returns the resolved irish.dict path after Ensure (may be empty).
func IrishPOSDictPath() string {
	EnsureDefaultIrishTagger()
	return gaPOSDictPath
}

func loadGAManualTagger(dictPath string) tagging.WordTagger {
	if wt := languagetool.LoadManualTaggerBesideDict(dictPath, []string{"added.txt", "added_custom.txt"}); wt != nil {
		return wt
	}
	if p := walkUpGAResourceDir(); p != "" {
		return languagetool.LoadManualTaggerFromDirs([]string{p}, []string{"added.txt", "added_custom.txt"})
	}
	return nil
}

func loadGAManualRemoval(dictPath string) tagging.WordTagger {
	if wt := languagetool.LoadManualTaggerBesideDict(dictPath, []string{"removed.txt", "removed_custom.txt"}); wt != nil {
		return wt
	}
	if p := walkUpGAResourceDir(); p != "" {
		return languagetool.LoadManualTaggerFromDirs([]string{p}, []string{"removed.txt", "removed_custom.txt"})
	}
	return nil
}

func walkUpGAResourceDir() string {
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "ga",
		"src", "main", "resources", "org", "languagetool", "resource", "ga")
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
