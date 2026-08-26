package uk

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging"
)

var (
	ukTaggerOnce  sync.Once
	ukPOSDictPath string
	// DefaultUkrainianTagger is the process-wide UkrainianTagger backed by
	// official /uk/ukrainian.dict (Maven morfologik-ukrainian-lt). Nil if missing.
	DefaultUkrainianTagger *UkrainianTagger
)

// DiscoverUkrainianPOSDict finds ukrainian.dict (Java resource /uk/ukrainian.dict).
// Order: LANG_UKRAINIAN_DICT, walk-up inspiration path, /tmp/uk-dict-fetch extract.
func DiscoverUkrainianPOSDict() string {
	if p := os.Getenv("LANG_UKRAINIAN_DICT"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	candidates := []string{
		filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "uk",
			"src", "main", "resources", "org", "languagetool", "resource", "uk", "ukrainian.dict"),
		"/tmp/uk-dict-fetch/org/languagetool/resource/uk/ukrainian.dict",
	}
	wd, err := os.Getwd()
	if err != nil {
		// still try absolute tmp
		if st, err := os.Stat(candidates[1]); err == nil && st.Mode().IsRegular() {
			return candidates[1]
		}
		return ""
	}
	dir := wd
	for i := 0; i < 14; i++ {
		for _, rel := range candidates {
			if filepath.IsAbs(rel) {
				if st, err := os.Stat(rel); err == nil && st.Mode().IsRegular() {
					return rel
				}
				continue
			}
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

// EnsureDefaultUkrainianTagger loads ukrainian.dict + official added/removed into
// DefaultUkrainianTagger. Idempotent; no-op if dict missing (fail closed).
func EnsureDefaultUkrainianTagger() {
	ukTaggerOnce.Do(func() {
		p := DiscoverUkrainianPOSDict()
		if p == "" {
			return
		}
		ukPOSDictPath = p
		mt := tagging.OpenMorfologikTagger(p)
		if mt == nil {
			return
		}
		wt := tagging.WordTagger(mt)
		// Java BaseTagger.initWordTagger: CombiningTagger when added.txt present
		// Official added/removed live in the language module resource tree (not jar).
		inspDirs := discoverUKResourceDirs(p)
		if manual := languagetool.LoadManualTaggerFromDirs(inspDirs, []string{"added.txt", "added_custom.txt"}); manual != nil {
			removal := languagetool.LoadManualTaggerFromDirs(inspDirs, []string{"removed.txt", "removed_custom.txt"})
			wt = tagging.NewCombiningTaggerWithRemoval(mt, manual, removal, false)
		} else if manual := languagetool.LoadManualTaggerBesideDict(p, []string{"added.txt", "added_custom.txt"}); manual != nil {
			removal := languagetool.LoadManualTaggerBesideDict(p, []string{"removed.txt", "removed_custom.txt"})
			wt = tagging.NewCombiningTaggerWithRemoval(mt, manual, removal, false)
		}
		DefaultUkrainianTagger = NewUkrainianTagger(wt)
	})
}

// UkrainianPOSDictPath returns the resolved ukrainian.dict path after Ensure (may be empty).
func UkrainianPOSDictPath() string {
	EnsureDefaultUkrainianTagger()
	return ukPOSDictPath
}

func discoverUKResourceDirs(dictPath string) []string {
	// Prefer inspiration module resource/uk (added.txt lives there).
	var dirs []string
	wd, err := os.Getwd()
	if err == nil {
		dir := wd
		for i := 0; i < 14; i++ {
			cand := filepath.Join(dir, "inspiration", "languagetool", "languagetool-language-modules", "uk",
				"src", "main", "resources", "org", "languagetool", "resource", "uk")
			if st, err := os.Stat(cand); err == nil && st.IsDir() {
				dirs = append(dirs, cand)
				break
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	// Also dir beside dict (jar extract layout)
	if dictPath != "" {
		dirs = append(dirs, filepath.Dir(dictPath))
	}
	return dirs
}
