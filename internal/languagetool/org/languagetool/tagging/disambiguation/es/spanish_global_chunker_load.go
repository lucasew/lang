package es

import (
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// Spanish GlobalChunker settings match Java SpanishHybridDisambiguator:
//
//	MultiWordChunker.getInstance("/spelling_global.txt", false, true, false, "NPCN000");
//	// constructor does NOT call setIgnoreSpelling (unlike NL GlobalChunker)
//	// constructor does NOT call setRemovePreviousTags on chunkerGlobal
//
// Differs from Spanish multiwords: allowFirstCapitalized is false; DefaultTag is "NPCN000"
// (not empty, not tagForNotAddingTags) → open/close <NPCN000></NPCN000> tags.
var spanishGlobalChunkerSettings = disambiguation.MultiWordChunkerSettings{
	AllowFirstCapitalized: false,
	AllowAllUppercase:     true,
	AllowTitlecase:        false,
	DefaultTag:            "NPCN000",
}

var (
	esGlobalChunkerOnce sync.Once
	esGlobalChunkerInst *disambiguation.MultiWordChunker
)

// OpenSpanishGlobalChunker ports MultiWordChunker.getInstance for
// /spelling_global.txt with SpanishHybridDisambiguator constructor settings
// (allowFirstCapitalized=false, DefaultTag "NPCN000"). Does not set ignore-spelling.
func OpenSpanishGlobalChunker(r io.Reader) (*disambiguation.MultiWordChunker, error) {
	return disambiguation.NewMultiWordChunkerFromReader(r, spanishGlobalChunkerSettings)
}

// LoadSpanishGlobalChunkerFromPath opens the official spelling_global file
// and builds MultiWordChunker with SpanishHybridDisambiguator GlobalChunker defaults.
func LoadSpanishGlobalChunkerFromPath(path string) (*disambiguation.MultiWordChunker, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return OpenSpanishGlobalChunker(f)
}

// SpanishGlobalChunker returns the process-cached MultiWordChunker for
// official /spelling_global.txt (Java SpanishHybridDisambiguator.chunkerGlobal field).
// Nil if the official resource is not discoverable.
func SpanishGlobalChunker() *disambiguation.MultiWordChunker {
	esGlobalChunkerOnce.Do(func() {
		p := DiscoverSpanishGlobalChunker()
		if p == "" {
			return
		}
		c, err := LoadSpanishGlobalChunkerFromPath(p)
		if err != nil || c == nil {
			return
		}
		esGlobalChunkerInst = c
	})
	return esGlobalChunkerInst
}

// DiscoverSpanishGlobalChunker finds official spelling_global.txt
// (Java resource /spelling_global.txt used by SpanishHybridDisambiguator.chunkerGlobal).
func DiscoverSpanishGlobalChunker() string {
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
