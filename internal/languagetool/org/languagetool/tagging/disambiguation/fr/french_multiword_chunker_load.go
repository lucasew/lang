package fr

import (
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// French multiwords settings match Java FrenchHybridDisambiguator:
//
//	MultiWordChunker.getInstance("/fr/multiwords.txt", true, true, false);
//	// default tag: null (normal multiword open-close tags — not tagForNotAddingTags)
//	chunker.setRemovePreviousTags(true);
//	// NO setIgnoreSpelling (unlike chunkerGlobal / NL)
var frenchMultiWordChunkerSettings = disambiguation.MultiWordChunkerSettings{
	AllowFirstCapitalized: true,
	AllowAllUppercase:     true,
	AllowTitlecase:        false,
	// DefaultTag empty: phrase;tag / phrase\ttag lines from official multiwords.txt
}

var (
	frMultiWordChunkerOnce sync.Once
	frMultiWordChunkerInst *disambiguation.MultiWordChunker
)

// OpenFrenchMultiWordChunker ports MultiWordChunker.getInstance for
// /fr/multiwords.txt with FrenchHybridDisambiguator constructor settings
// and setRemovePreviousTags(true). Does not set ignore-spelling.
func OpenFrenchMultiWordChunker(r io.Reader) (*disambiguation.MultiWordChunker, error) {
	c, err := disambiguation.NewMultiWordChunkerFromReader(r, frenchMultiWordChunkerSettings)
	if err != nil {
		return nil, err
	}
	c.SetRemovePreviousTags(true)
	return c, nil
}

// LoadFrenchMultiWordChunkerFromPath opens the official multiwords file
// and builds MultiWordChunker with FrenchHybridDisambiguator multiwords defaults.
func LoadFrenchMultiWordChunkerFromPath(path string) (*disambiguation.MultiWordChunker, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return OpenFrenchMultiWordChunker(f)
}

// FrenchMultiWordChunker returns the process-cached MultiWordChunker for
// official /fr/multiwords.txt (Java FrenchHybridDisambiguator.chunker field).
// Nil if the official resource is not discoverable.
func FrenchMultiWordChunker() *disambiguation.MultiWordChunker {
	frMultiWordChunkerOnce.Do(func() {
		p := DiscoverFrenchMultiwords()
		if p == "" {
			return
		}
		c, err := LoadFrenchMultiWordChunkerFromPath(p)
		if err != nil || c == nil {
			return
		}
		frMultiWordChunkerInst = c
	})
	return frMultiWordChunkerInst
}

// DiscoverFrenchMultiwords finds official fr/multiwords.txt
// (Java resource /fr/multiwords.txt used by FrenchHybridDisambiguator.chunker).
func DiscoverFrenchMultiwords() string {
	if p := os.Getenv("LANG_FR_MULTIWORDS_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "fr",
		"src", "main", "resources", "org", "languagetool", "resource", "fr", "multiwords.txt")
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
