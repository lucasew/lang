package sr

import (
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// Serbian multiwords settings match Java SerbianHybridDisambiguator:
//
//	// historical: new MultiWordChunker("/sr/multiwords.txt")
//	// match MultiWordChunker.getInstance defaults: false,false,false
//	// default tag: null (normal multiword open-close tags — not tagForNotAddingTags)
//	// NO setRemovePreviousTags, NO setIgnoreSpelling
//
// Official sr/multiwords.txt is currently empty (0 lines).
// Java still constructs MultiWordChunker — empty maps after load is correct
// (not "skip stage"). Do not invent multiword entries.
var serbianMultiWordChunkerSettings = disambiguation.MultiWordChunkerSettings{
	AllowFirstCapitalized: false,
	AllowAllUppercase:     false,
	AllowTitlecase:        false,
	// DefaultTag empty: phrase\ttag lines from official multiwords.txt (none today)
}

var (
	srMultiWordChunkerOnce sync.Once
	srMultiWordChunkerInst *disambiguation.MultiWordChunker
)

// OpenSerbianMultiWordChunker ports MultiWordChunker for
// /sr/multiwords.txt with SerbianHybridDisambiguator constructor defaults.
// Does not set remove-previous-tags or ignore-spelling.
func OpenSerbianMultiWordChunker(r io.Reader) (*disambiguation.MultiWordChunker, error) {
	return disambiguation.NewMultiWordChunkerFromReader(r, serbianMultiWordChunkerSettings)
}

// LoadSerbianMultiWordChunkerFromPath opens the official multiwords file
// and builds MultiWordChunker with SerbianHybridDisambiguator multiwords defaults.
func LoadSerbianMultiWordChunkerFromPath(path string) (*disambiguation.MultiWordChunker, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return OpenSerbianMultiWordChunker(f)
}

// SerbianMultiWordChunker returns the process-cached MultiWordChunker for
// official /sr/multiwords.txt (Java SerbianHybridDisambiguator.chunker field).
// Nil if the official resource is not discoverable.
// Official file may be empty → non-nil chunker with empty phrase maps.
func SerbianMultiWordChunker() *disambiguation.MultiWordChunker {
	srMultiWordChunkerOnce.Do(func() {
		p := DiscoverSerbianMultiwords()
		if p == "" {
			return
		}
		c, err := LoadSerbianMultiWordChunkerFromPath(p)
		if err != nil || c == nil {
			return
		}
		srMultiWordChunkerInst = c
	})
	return srMultiWordChunkerInst
}

// DiscoverSerbianMultiwords finds official sr/multiwords.txt
// (Java resource /sr/multiwords.txt used by SerbianHybridDisambiguator.chunker).
func DiscoverSerbianMultiwords() string {
	if p := os.Getenv("LANG_SR_MULTIWORDS_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "sr",
		"src", "main", "resources", "org", "languagetool", "resource", "sr", "multiwords.txt")
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
