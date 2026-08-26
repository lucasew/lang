package tl

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

var (
	tlTaggerOnce  sync.Once
	tlPOSDictPath string
)

// DefaultTagalogTagger is the process singleton (Java new TagalogTagger()).
// Backed by MapWordTagger until EnsureDefaultTagalogTagger loads tagalog.dict.
var DefaultTagalogTagger = NewTagalogTagger(tagging.MapWordTagger{})

// DiscoverTagalogPOSDict finds tagalog.dict (Java resource /tl/tagalog.dict).
// Order: LANG_TAGALOG_DICT, walk-up inspiration module path.
func DiscoverTagalogPOSDict() string {
	if p := os.Getenv("LANG_TAGALOG_DICT"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	relPaths := []string{
		filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "tl",
			"src", "main", "resources", "org", "languagetool", "resource", "tl", "tagalog.dict"),
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

// EnsureDefaultTagalogTagger loads tagalog.dict into DefaultTagalogTagger.
// Idempotent; no-op if dict missing (fail closed).
func EnsureDefaultTagalogTagger() {
	tlTaggerOnce.Do(func() {
		p := DiscoverTagalogPOSDict()
		if p == "" {
			return
		}
		tlPOSDictPath = p
		mt := tagging.OpenMorfologikTagger(p)
		if mt == nil {
			return
		}
		// Java BaseTagger.initWordTagger: CombiningTagger when added.txt present
		wt := tagging.WordTagger(mt)
		if manual := loadTLManualTagger(p); manual != nil {
			removal := loadTLManualRemoval(p)
			wt = tagging.NewCombiningTaggerWithRemoval(wt, manual, removal, false)
		}
		DefaultTagalogTagger = NewTagalogTagger(wt)
	})
}

// TagalogPOSDictPath returns the resolved tagalog.dict path after Ensure (may be empty).
func TagalogPOSDictPath() string {
	EnsureDefaultTagalogTagger()
	return tlPOSDictPath
}

func loadTLManualTagger(dictPath string) tagging.WordTagger {
	if wt := languagetool.LoadManualTaggerBesideDict(dictPath, []string{"added.txt", "added_custom.txt"}); wt != nil {
		return wt
	}
	if p := walkUpTLResourceDir(); p != "" {
		return languagetool.LoadManualTaggerFromDirs([]string{p}, []string{"added.txt", "added_custom.txt"})
	}
	return nil
}

func loadTLManualRemoval(dictPath string) tagging.WordTagger {
	if wt := languagetool.LoadManualTaggerBesideDict(dictPath, []string{"removed.txt", "removed_custom.txt"}); wt != nil {
		return wt
	}
	if p := walkUpTLResourceDir(); p != "" {
		return languagetool.LoadManualTaggerFromDirs([]string{p}, []string{"removed.txt", "removed_custom.txt"})
	}
	return nil
}

func walkUpTLResourceDir() string {
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "tl",
		"src", "main", "resources", "org", "languagetool", "resource", "tl")
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
