package ga

import (
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// Irish multiwords settings match Java IrishHybridDisambiguator:
//
//	MultiWordChunker.getInstance("/ga/multiwords.txt");
//	// defaults: allowFirstCapitalized=false, allowAllUppercase=false, allowTitlecase=false
//	// default tag: null (normal multiword open-close tags — not tagForNotAddingTags)
//	// NO setRemovePreviousTags, NO setIgnoreSpelling
var irishMultiWordChunkerSettings = disambiguation.MultiWordChunkerSettings{
	AllowFirstCapitalized: false,
	AllowAllUppercase:     false,
	AllowTitlecase:        false,
	// DefaultTag empty: phrase\ttag lines from official multiwords.txt
}

var (
	gaMultiWordChunkerOnce sync.Once
	gaMultiWordChunkerInst *disambiguation.MultiWordChunker
)

// OpenIrishMultiWordChunker ports MultiWordChunker.getInstance for
// /ga/multiwords.txt with IrishHybridDisambiguator constructor defaults.
// Does not set remove-previous-tags or ignore-spelling.
func OpenIrishMultiWordChunker(r io.Reader) (*disambiguation.MultiWordChunker, error) {
	return disambiguation.NewMultiWordChunkerFromReader(r, irishMultiWordChunkerSettings)
}

// LoadIrishMultiWordChunkerFromPath opens the official multiwords file
// and builds MultiWordChunker with IrishHybridDisambiguator multiwords defaults.
func LoadIrishMultiWordChunkerFromPath(path string) (*disambiguation.MultiWordChunker, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return OpenIrishMultiWordChunker(f)
}

// IrishMultiWordChunker returns the process-cached MultiWordChunker for
// official /ga/multiwords.txt (Java IrishHybridDisambiguator.chunker field).
// Nil if the official resource is not discoverable.
func IrishMultiWordChunker() *disambiguation.MultiWordChunker {
	gaMultiWordChunkerOnce.Do(func() {
		p := DiscoverIrishMultiwords()
		if p == "" {
			return
		}
		c, err := LoadIrishMultiWordChunkerFromPath(p)
		if err != nil || c == nil {
			return
		}
		gaMultiWordChunkerInst = c
	})
	return gaMultiWordChunkerInst
}

// DiscoverIrishMultiwords finds official ga/multiwords.txt
// (Java resource /ga/multiwords.txt used by IrishHybridDisambiguator.chunker).
func DiscoverIrishMultiwords() string {
	if p := os.Getenv("LANG_GA_MULTIWORDS_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "ga",
		"src", "main", "resources", "org", "languagetool", "resource", "ga", "multiwords.txt")
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
