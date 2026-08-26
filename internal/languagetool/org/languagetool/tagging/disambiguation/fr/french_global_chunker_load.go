package fr

import (
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// French GlobalChunker settings match Java FrenchHybridDisambiguator:
//
//	MultiWordChunker.getInstance("/spelling_global.txt", false, true, false,
//	  MultiWordChunker.tagForNotAddingTags);
//	chunkerGlobal.setIgnoreSpelling(true);
//
// Differs from French multiwords: allowFirstCapitalized is false (not true);
// DefaultTag is tagForNotAddingTags; setIgnoreSpelling(true) (multiwords does neither).
// Same settings as NL DutchHybridDisambiguator.chunkerGlobal / DE MultitokenGlobal.
var frenchGlobalChunkerSettings = disambiguation.MultiWordChunkerSettings{
	AllowFirstCapitalized: false,
	AllowAllUppercase:     true,
	AllowTitlecase:        false,
	DefaultTag:            disambiguation.TagForNotAddingTags,
}

var (
	frGlobalChunkerOnce sync.Once
	frGlobalChunkerInst *disambiguation.MultiWordChunker
)

// OpenFrenchGlobalChunker ports MultiWordChunker.getInstance for
// /spelling_global.txt with FrenchHybridDisambiguator constructor settings
// (allowFirstCapitalized=false) and setIgnoreSpelling(true).
func OpenFrenchGlobalChunker(r io.Reader) (*disambiguation.MultiWordChunker, error) {
	c, err := disambiguation.NewMultiWordChunkerFromReader(r, frenchGlobalChunkerSettings)
	if err != nil {
		return nil, err
	}
	c.SetIgnoreSpelling(true)
	return c, nil
}

// LoadFrenchGlobalChunkerFromPath opens the official spelling_global file
// and builds MultiWordChunker with FrenchHybridDisambiguator GlobalChunker defaults.
func LoadFrenchGlobalChunkerFromPath(path string) (*disambiguation.MultiWordChunker, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return OpenFrenchGlobalChunker(f)
}

// FrenchGlobalChunker returns the process-cached MultiWordChunker for
// official /spelling_global.txt (Java FrenchHybridDisambiguator.chunkerGlobal field).
// Nil if the official resource is not discoverable.
func FrenchGlobalChunker() *disambiguation.MultiWordChunker {
	frGlobalChunkerOnce.Do(func() {
		p := DiscoverFrenchGlobalChunker()
		if p == "" {
			return
		}
		c, err := LoadFrenchGlobalChunkerFromPath(p)
		if err != nil || c == nil {
			return
		}
		frGlobalChunkerInst = c
	})
	return frGlobalChunkerInst
}

// DiscoverFrenchGlobalChunker finds official spelling_global.txt
// (Java resource /spelling_global.txt used by FrenchHybridDisambiguator.chunkerGlobal).
func DiscoverFrenchGlobalChunker() string {
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
