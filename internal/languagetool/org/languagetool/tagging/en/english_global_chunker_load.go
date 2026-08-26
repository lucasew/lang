package en

import (
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// English GlobalChunker settings match Java EnglishHybridDisambiguator:
//
//	MultiWordChunker.getInstance("/spelling_global.txt", true, true, false,
//	  MultiWordChunker.tagForNotAddingTags);
//	chunkerGlobal.setIgnoreSpelling(true);
//	// NO setRemovePreviousTags on chunkerGlobal
//
// Differs from French/Dutch GlobalChunker: allowFirstCapitalized is true (not false).
// Differs from English multiwords: DefaultTag is tagForNotAddingTags; no setRemovePreviousTags.
var englishGlobalChunkerSettings = disambiguation.MultiWordChunkerSettings{
	AllowFirstCapitalized: true,
	AllowAllUppercase:     true,
	AllowTitlecase:        false,
	DefaultTag:            disambiguation.TagForNotAddingTags,
}

var (
	enGlobalChunkerOnce sync.Once
	enGlobalChunkerInst *disambiguation.MultiWordChunker
)

// OpenEnglishGlobalChunker ports MultiWordChunker.getInstance for
// /spelling_global.txt with EnglishHybridDisambiguator constructor settings
// (allowFirstCapitalized=true) and setIgnoreSpelling(true).
func OpenEnglishGlobalChunker(r io.Reader) (*disambiguation.MultiWordChunker, error) {
	c, err := disambiguation.NewMultiWordChunkerFromReader(r, englishGlobalChunkerSettings)
	if err != nil {
		return nil, err
	}
	c.SetIgnoreSpelling(true)
	return c, nil
}

// LoadEnglishGlobalChunkerFromPath opens the official spelling_global file
// and builds MultiWordChunker with EnglishHybridDisambiguator GlobalChunker defaults.
func LoadEnglishGlobalChunkerFromPath(path string) (*disambiguation.MultiWordChunker, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return OpenEnglishGlobalChunker(f)
}

// EnglishGlobalChunker returns the process-cached MultiWordChunker for
// official /spelling_global.txt (Java EnglishHybridDisambiguator.chunkerGlobal field).
// Nil if the official resource is not discoverable.
func EnglishGlobalChunker() *disambiguation.MultiWordChunker {
	enGlobalChunkerOnce.Do(func() {
		p := DiscoverEnglishGlobalChunker()
		if p == "" {
			return
		}
		c, err := LoadEnglishGlobalChunkerFromPath(p)
		if err != nil || c == nil {
			return
		}
		enGlobalChunkerInst = c
	})
	return enGlobalChunkerInst
}

// DiscoverEnglishGlobalChunker finds official spelling_global.txt
// (Java resource /spelling_global.txt used by EnglishHybridDisambiguator.chunkerGlobal).
func DiscoverEnglishGlobalChunker() string {
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
