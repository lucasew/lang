package ru

import (
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// Russian multiwords settings match Java RussianHybridDisambiguator:
//
//	MultiWordChunker.getInstance("/ru/multiwords.txt");
//	// defaults: allowFirstCapitalized=false, allowAllUppercase=false, allowTitlecase=false
//	// default tag: null (normal multiword open-close tags — not tagForNotAddingTags)
//	// NO setRemovePreviousTags, NO setIgnoreSpelling
var russianMultiWordChunkerSettings = disambiguation.MultiWordChunkerSettings{
	AllowFirstCapitalized: false,
	AllowAllUppercase:     false,
	AllowTitlecase:        false,
	// DefaultTag empty: phrase\ttag lines from official multiwords.txt
}

var (
	ruMultiWordChunkerOnce sync.Once
	ruMultiWordChunkerInst *disambiguation.MultiWordChunker
)

// OpenRussianMultiWordChunker ports MultiWordChunker.getInstance for
// /ru/multiwords.txt with RussianHybridDisambiguator constructor defaults.
// Does not set remove-previous-tags or ignore-spelling.
func OpenRussianMultiWordChunker(r io.Reader) (*disambiguation.MultiWordChunker, error) {
	return disambiguation.NewMultiWordChunkerFromReader(r, russianMultiWordChunkerSettings)
}

// LoadRussianMultiWordChunkerFromPath opens the official multiwords file
// and builds MultiWordChunker with RussianHybridDisambiguator multiwords defaults.
func LoadRussianMultiWordChunkerFromPath(path string) (*disambiguation.MultiWordChunker, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return OpenRussianMultiWordChunker(f)
}

// RussianMultiWordChunker returns the process-cached MultiWordChunker for
// official /ru/multiwords.txt (Java RussianHybridDisambiguator.chunker field).
// Nil if the official resource is not discoverable.
func RussianMultiWordChunker() *disambiguation.MultiWordChunker {
	ruMultiWordChunkerOnce.Do(func() {
		p := DiscoverRussianMultiwords()
		if p == "" {
			return
		}
		c, err := LoadRussianMultiWordChunkerFromPath(p)
		if err != nil || c == nil {
			return
		}
		ruMultiWordChunkerInst = c
	})
	return ruMultiWordChunkerInst
}

// DiscoverRussianMultiwords finds official ru/multiwords.txt
// (Java resource /ru/multiwords.txt used by RussianHybridDisambiguator.chunker).
func DiscoverRussianMultiwords() string {
	if p := os.Getenv("LANG_RU_MULTIWORDS_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "ru",
		"src", "main", "resources", "org", "languagetool", "resource", "ru", "multiwords.txt")
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
