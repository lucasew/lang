package ca

import (
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// Catalan GlobalChunker settings match Java CatalanHybridDisambiguator:
//
//	MultiWordChunker.getInstance("/spelling_global.txt", false, true, false, "NPCN000");
//	// constructor does NOT call setIgnoreSpelling (unlike NL GlobalChunker)
//	// constructor does NOT call setRemovePreviousTags on chunkerGlobal
//
// Differs from Catalan multiwords: allowFirstCapitalized is false; DefaultTag is "NPCN000"
// (not empty, not tagForNotAddingTags) → open/close <NPCN000></NPCN000> tags.
var catalanGlobalChunkerSettings = disambiguation.MultiWordChunkerSettings{
	AllowFirstCapitalized: false,
	AllowAllUppercase:     true,
	AllowTitlecase:        false,
	DefaultTag:            "NPCN000",
}

var (
	caGlobalChunkerOnce sync.Once
	caGlobalChunkerInst *disambiguation.MultiWordChunker
)

// OpenCatalanGlobalChunker ports MultiWordChunker.getInstance for
// /spelling_global.txt with CatalanHybridDisambiguator constructor settings
// (allowFirstCapitalized=false, DefaultTag "NPCN000"). Does not set ignore-spelling.
func OpenCatalanGlobalChunker(r io.Reader) (*disambiguation.MultiWordChunker, error) {
	return disambiguation.NewMultiWordChunkerFromReader(r, catalanGlobalChunkerSettings)
}

// LoadCatalanGlobalChunkerFromPath opens the official spelling_global file
// and builds MultiWordChunker with CatalanHybridDisambiguator GlobalChunker defaults.
func LoadCatalanGlobalChunkerFromPath(path string) (*disambiguation.MultiWordChunker, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return OpenCatalanGlobalChunker(f)
}

// CatalanGlobalChunker returns the process-cached MultiWordChunker for
// official /spelling_global.txt (Java CatalanHybridDisambiguator.chunkerGlobal field).
// Nil if the official resource is not discoverable.
func CatalanGlobalChunker() *disambiguation.MultiWordChunker {
	caGlobalChunkerOnce.Do(func() {
		p := DiscoverCatalanGlobalChunker()
		if p == "" {
			return
		}
		c, err := LoadCatalanGlobalChunkerFromPath(p)
		if err != nil || c == nil {
			return
		}
		caGlobalChunkerInst = c
	})
	return caGlobalChunkerInst
}

// DiscoverCatalanGlobalChunker finds official spelling_global.txt
// (Java resource /spelling_global.txt used by CatalanHybridDisambiguator.chunkerGlobal).
func DiscoverCatalanGlobalChunker() string {
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
