package de

import (
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// German MultitokenGlobal settings match Java GermanRuleDisambiguator:
//
//	MultiWordChunker.getInstance("/spelling_global.txt", false, true, false,
//	  MultiWordChunker.tagForNotAddingTags);
//	multitokenSpeller3.setIgnoreSpelling(true);
//
// Differs from MultitokenIgnore: allowFirstCapitalized is false (not true).
var germanMultitokenGlobalSettings = disambiguation.MultiWordChunkerSettings{
	AllowFirstCapitalized: false,
	AllowAllUppercase:     true,
	AllowTitlecase:        false,
	DefaultTag:            disambiguation.TagForNotAddingTags,
}

var (
	deMultitokenGlobalOnce sync.Once
	deMultitokenGlobalInst *disambiguation.MultiWordChunker
)

// OpenGermanMultitokenGlobal ports MultiWordChunker.getInstance for
// /spelling_global.txt with GermanRuleDisambiguator constructor settings
// (allowFirstCapitalized=false) and setIgnoreSpelling(true).
func OpenGermanMultitokenGlobal(r io.Reader) (*disambiguation.MultiWordChunker, error) {
	c, err := disambiguation.NewMultiWordChunkerFromReader(r, germanMultitokenGlobalSettings)
	if err != nil {
		return nil, err
	}
	c.SetIgnoreSpelling(true)
	return c, nil
}

// LoadGermanMultitokenGlobalFromPath opens the official spelling_global file
// and builds MultiWordChunker with GermanRuleDisambiguator MultitokenGlobal defaults.
func LoadGermanMultitokenGlobalFromPath(path string) (*disambiguation.MultiWordChunker, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return OpenGermanMultitokenGlobal(f)
}

// GermanMultitokenGlobal returns the process-cached MultiWordChunker for
// official /spelling_global.txt (Java multitokenSpeller3 field).
// Nil if the official resource is not discoverable.
func GermanMultitokenGlobal() *disambiguation.MultiWordChunker {
	deMultitokenGlobalOnce.Do(func() {
		p := DiscoverGermanMultitokenGlobal()
		if p == "" {
			return
		}
		c, err := LoadGermanMultitokenGlobalFromPath(p)
		if err != nil || c == nil {
			return
		}
		deMultitokenGlobalInst = c
	})
	return deMultitokenGlobalInst
}

// DiscoverGermanMultitokenGlobal finds official spelling_global.txt
// (Java resource /spelling_global.txt used by GermanRuleDisambiguator.multitokenSpeller3).
func DiscoverGermanMultitokenGlobal() string {
	if p := os.Getenv("LANG_SPELLING_GLOBAL_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-core",
		"src", "main", "resources", "org", "languagetool", "resource", "spelling_global.txt")
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	dir := wd
	for i := 0; i < 14; i++ {
		p := filepath.Join(dir, rel)
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
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
