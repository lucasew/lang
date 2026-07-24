package sv

import (
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// Swedish multiwords settings match Java SwedishHybridDisambiguator:
//
//	MultiWordChunker.getInstance("/sv/multiwords.txt");
//	// defaults: allowFirstCapitalized=false, allowAllUppercase=false, allowTitlecase=false
//	// default tag: null (normal multiword open-close tags — not tagForNotAddingTags)
//	// NO setRemovePreviousTags, NO setIgnoreSpelling
var swedishMultiWordChunkerSettings = disambiguation.MultiWordChunkerSettings{
	AllowFirstCapitalized: false,
	AllowAllUppercase:     false,
	AllowTitlecase:        false,
	// DefaultTag empty: phrase\ttag lines from official multiwords.txt
}

var (
	svMultiWordChunkerOnce sync.Once
	svMultiWordChunkerInst *disambiguation.MultiWordChunker
)

// OpenSwedishMultiWordChunker ports MultiWordChunker.getInstance for
// /sv/multiwords.txt with SwedishHybridDisambiguator constructor defaults.
// Does not set remove-previous-tags or ignore-spelling.
func OpenSwedishMultiWordChunker(r io.Reader) (*disambiguation.MultiWordChunker, error) {
	return disambiguation.NewMultiWordChunkerFromReader(r, swedishMultiWordChunkerSettings)
}

// LoadSwedishMultiWordChunkerFromPath opens the official multiwords file
// and builds MultiWordChunker with SwedishHybridDisambiguator multiwords defaults.
func LoadSwedishMultiWordChunkerFromPath(path string) (*disambiguation.MultiWordChunker, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return OpenSwedishMultiWordChunker(f)
}

// SwedishMultiWordChunker returns the process-cached MultiWordChunker for
// official /sv/multiwords.txt (Java SwedishHybridDisambiguator.chunker field).
// Nil if the official resource is not discoverable.
func SwedishMultiWordChunker() *disambiguation.MultiWordChunker {
	svMultiWordChunkerOnce.Do(func() {
		p := DiscoverSwedishMultiwords()
		if p == "" {
			return
		}
		c, err := LoadSwedishMultiWordChunkerFromPath(p)
		if err != nil || c == nil {
			return
		}
		svMultiWordChunkerInst = c
	})
	return svMultiWordChunkerInst
}

// DiscoverSwedishMultiwords finds official sv/multiwords.txt
// (Java resource /sv/multiwords.txt used by SwedishHybridDisambiguator.chunker).
func DiscoverSwedishMultiwords() string {
	if p := os.Getenv("LANG_SV_MULTIWORDS_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "sv",
		"src", "main", "resources", "org", "languagetool", "resource", "sv", "multiwords.txt")
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
