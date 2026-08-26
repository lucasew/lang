package en

import (
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// English multiwords settings match Java EnglishHybridDisambiguator:
//
//	MultiWordChunker.getInstance("/en/multiwords.txt", true, true, false);
//	// default tag: null (normal multiword open-close tags — not tagForNotAddingTags)
//	chunker.setIgnoreSpelling(true);
//	chunker.setRemovePreviousTags(true);
var englishMultiWordChunkerSettings = disambiguation.MultiWordChunkerSettings{
	AllowFirstCapitalized: true,
	AllowAllUppercase:     true,
	AllowTitlecase:        false,
	// DefaultTag empty: phrase\ttag lines from official multiwords.txt
}

var (
	enMultiWordChunkerOnce sync.Once
	enMultiWordChunkerInst *disambiguation.MultiWordChunker
)

// OpenEnglishMultiWordChunker ports MultiWordChunker.getInstance for
// /en/multiwords.txt with EnglishHybridDisambiguator constructor settings,
// setIgnoreSpelling(true), and setRemovePreviousTags(true).
func OpenEnglishMultiWordChunker(r io.Reader) (*disambiguation.MultiWordChunker, error) {
	c, err := disambiguation.NewMultiWordChunkerFromReader(r, englishMultiWordChunkerSettings)
	if err != nil {
		return nil, err
	}
	c.SetIgnoreSpelling(true)
	c.SetRemovePreviousTags(true)
	return c, nil
}

// LoadEnglishMultiWordChunkerFromPath opens the official multiwords file
// and builds MultiWordChunker with EnglishHybridDisambiguator multiwords defaults.
func LoadEnglishMultiWordChunkerFromPath(path string) (*disambiguation.MultiWordChunker, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return OpenEnglishMultiWordChunker(f)
}

// EnglishMultiWordChunker returns the process-cached MultiWordChunker for
// official /en/multiwords.txt (Java EnglishHybridDisambiguator.chunker field).
// Nil if the official resource is not discoverable.
func EnglishMultiWordChunker() *disambiguation.MultiWordChunker {
	enMultiWordChunkerOnce.Do(func() {
		p := DiscoverEnglishMultiwords()
		if p == "" {
			return
		}
		c, err := LoadEnglishMultiWordChunkerFromPath(p)
		if err != nil || c == nil {
			return
		}
		enMultiWordChunkerInst = c
	})
	return enMultiWordChunkerInst
}

// DiscoverEnglishMultiwords finds official en/multiwords.txt
// (Java resource /en/multiwords.txt used by EnglishHybridDisambiguator.chunker).
func DiscoverEnglishMultiwords() string {
	if p := os.Getenv("LANG_EN_MULTIWORDS_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	// Also honor LANG_EN_MULTIWORDS (commandline discover twin env).
	if p := os.Getenv("LANG_EN_MULTIWORDS"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "en",
		"src", "main", "resources", "org", "languagetool", "resource", "en", "multiwords.txt")
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
