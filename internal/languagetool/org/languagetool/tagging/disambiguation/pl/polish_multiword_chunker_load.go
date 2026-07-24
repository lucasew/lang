package pl

import (
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// Polish multiwords settings match Java PolishHybridDisambiguator:
//
//	MultiWordChunker.getInstance("/pl/multiwords.txt");
//	// defaults: allowFirstCapitalized=false, allowAllUppercase=false, allowTitlecase=false
//	// default tag: null (normal multiword open-close tags — not tagForNotAddingTags)
//	// NO setRemovePreviousTags, NO setIgnoreSpelling
var polishMultiWordChunkerSettings = disambiguation.MultiWordChunkerSettings{
	AllowFirstCapitalized: false,
	AllowAllUppercase:     false,
	AllowTitlecase:        false,
	// DefaultTag empty: phrase\ttag lines from official multiwords.txt
}

var (
	plMultiWordChunkerOnce sync.Once
	plMultiWordChunkerInst *disambiguation.MultiWordChunker
)

// OpenPolishMultiWordChunker ports MultiWordChunker.getInstance for
// /pl/multiwords.txt with PolishHybridDisambiguator constructor defaults.
// Does not set remove-previous-tags or ignore-spelling.
func OpenPolishMultiWordChunker(r io.Reader) (*disambiguation.MultiWordChunker, error) {
	return disambiguation.NewMultiWordChunkerFromReader(r, polishMultiWordChunkerSettings)
}

// LoadPolishMultiWordChunkerFromPath opens the official multiwords file
// and builds MultiWordChunker with PolishHybridDisambiguator multiwords defaults.
func LoadPolishMultiWordChunkerFromPath(path string) (*disambiguation.MultiWordChunker, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return OpenPolishMultiWordChunker(f)
}

// PolishMultiWordChunker returns the process-cached MultiWordChunker for
// official /pl/multiwords.txt (Java PolishHybridDisambiguator.chunker field).
// Nil if the official resource is not discoverable.
func PolishMultiWordChunker() *disambiguation.MultiWordChunker {
	plMultiWordChunkerOnce.Do(func() {
		p := DiscoverPolishMultiwords()
		if p == "" {
			return
		}
		c, err := LoadPolishMultiWordChunkerFromPath(p)
		if err != nil || c == nil {
			return
		}
		plMultiWordChunkerInst = c
	})
	return plMultiWordChunkerInst
}

// DiscoverPolishMultiwords finds official pl/multiwords.txt
// (Java resource /pl/multiwords.txt used by PolishHybridDisambiguator.chunker).
func DiscoverPolishMultiwords() string {
	if p := os.Getenv("LANG_PL_MULTIWORDS_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "pl",
		"src", "main", "resources", "org", "languagetool", "resource", "pl", "multiwords.txt")
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
