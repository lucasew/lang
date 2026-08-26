package ca

import (
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// Catalan multiwords settings match Java CatalanHybridDisambiguator:
//
//	MultiWordChunker.getInstance("/ca/multiwords.txt", true, true, false);
//	// default tag: null (normal multiword open-close tags — not tagForNotAddingTags)
//	chunker.setRemovePreviousTags(true);
//	// NO setIgnoreSpelling (unlike PT/NL)
var catalanMultiWordChunkerSettings = disambiguation.MultiWordChunkerSettings{
	AllowFirstCapitalized: true,
	AllowAllUppercase:     true,
	AllowTitlecase:        false,
	// DefaultTag empty: phrase;tag / phrase\ttag lines from official multiwords.txt
}

var (
	caMultiWordChunkerOnce sync.Once
	caMultiWordChunkerInst *disambiguation.MultiWordChunker
)

// OpenCatalanMultiWordChunker ports MultiWordChunker.getInstance for
// /ca/multiwords.txt with CatalanHybridDisambiguator constructor settings
// and setRemovePreviousTags(true). Does not set ignore-spelling.
func OpenCatalanMultiWordChunker(r io.Reader) (*disambiguation.MultiWordChunker, error) {
	c, err := disambiguation.NewMultiWordChunkerFromReader(r, catalanMultiWordChunkerSettings)
	if err != nil {
		return nil, err
	}
	c.SetRemovePreviousTags(true)
	return c, nil
}

// LoadCatalanMultiWordChunkerFromPath opens the official multiwords file
// and builds MultiWordChunker with CatalanHybridDisambiguator multiwords defaults.
func LoadCatalanMultiWordChunkerFromPath(path string) (*disambiguation.MultiWordChunker, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return OpenCatalanMultiWordChunker(f)
}

// CatalanMultiWordChunker returns the process-cached MultiWordChunker for
// official /ca/multiwords.txt (Java CatalanHybridDisambiguator.chunker field).
// Nil if the official resource is not discoverable.
func CatalanMultiWordChunker() *disambiguation.MultiWordChunker {
	caMultiWordChunkerOnce.Do(func() {
		p := DiscoverCatalanMultiwords()
		if p == "" {
			return
		}
		c, err := LoadCatalanMultiWordChunkerFromPath(p)
		if err != nil || c == nil {
			return
		}
		caMultiWordChunkerInst = c
	})
	return caMultiWordChunkerInst
}

// DiscoverCatalanMultiwords finds official ca/multiwords.txt
// (Java resource /ca/multiwords.txt used by CatalanHybridDisambiguator.chunker).
func DiscoverCatalanMultiwords() string {
	if p := os.Getenv("LANG_CA_MULTIWORDS_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "ca",
		"src", "main", "resources", "org", "languagetool", "resource", "ca", "multiwords.txt")
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
