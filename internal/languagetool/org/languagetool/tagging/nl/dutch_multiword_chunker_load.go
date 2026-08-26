package nl

import (
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// Dutch multiwords settings match Java DutchHybridDisambiguator:
//
//	MultiWordChunker.getInstance("/nl/multiwords.txt", true, true, false,
//	  MultiWordChunker.tagForNotAddingTags);
//	chunker.setIgnoreSpelling(true);
var dutchMultiWordChunkerSettings = disambiguation.MultiWordChunkerSettings{
	AllowFirstCapitalized: true,
	AllowAllUppercase:     true,
	AllowTitlecase:        false,
	DefaultTag:            disambiguation.TagForNotAddingTags,
}

var (
	nlMultiWordChunkerOnce sync.Once
	nlMultiWordChunkerInst *disambiguation.MultiWordChunker
)

// OpenDutchMultiWordChunker ports MultiWordChunker.getInstance for
// /nl/multiwords.txt with DutchHybridDisambiguator constructor settings
// and setIgnoreSpelling(true).
func OpenDutchMultiWordChunker(r io.Reader) (*disambiguation.MultiWordChunker, error) {
	c, err := disambiguation.NewMultiWordChunkerFromReader(r, dutchMultiWordChunkerSettings)
	if err != nil {
		return nil, err
	}
	c.SetIgnoreSpelling(true)
	return c, nil
}

// LoadDutchMultiWordChunkerFromPath opens the official multiwords file
// and builds MultiWordChunker with DutchHybridDisambiguator multiwords defaults.
func LoadDutchMultiWordChunkerFromPath(path string) (*disambiguation.MultiWordChunker, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return OpenDutchMultiWordChunker(f)
}

// DutchMultiWordChunker returns the process-cached MultiWordChunker for
// official /nl/multiwords.txt (Java DutchHybridDisambiguator.chunker field).
// Nil if the official resource is not discoverable.
func DutchMultiWordChunker() *disambiguation.MultiWordChunker {
	nlMultiWordChunkerOnce.Do(func() {
		p := DiscoverDutchMultiwords()
		if p == "" {
			return
		}
		c, err := LoadDutchMultiWordChunkerFromPath(p)
		if err != nil || c == nil {
			return
		}
		nlMultiWordChunkerInst = c
	})
	return nlMultiWordChunkerInst
}

// DiscoverDutchMultiwords finds official nl/multiwords.txt
// (Java resource /nl/multiwords.txt used by DutchHybridDisambiguator.chunker).
func DiscoverDutchMultiwords() string {
	if p := os.Getenv("LANG_NL_MULTIWORDS_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "nl",
		"src", "main", "resources", "org", "languagetool", "resource", "nl", "multiwords.txt")
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
