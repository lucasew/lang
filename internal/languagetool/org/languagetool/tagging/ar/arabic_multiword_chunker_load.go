package ar

import (
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// Arabic multiwords settings match Java ArabicHybridDisambiguator:
//
//	MultiWordChunker.getInstance("/ar/multiwords.txt");
//	// defaults: allowFirstCapitalized=false, allowAllUppercase=false, allowTitlecase=false
//	// default tag: null (normal multiword open-close tags — not tagForNotAddingTags)
//	// NO setRemovePreviousTags, NO setIgnoreSpelling
//
// Official ar/multiwords.txt is currently comment-only (no phrase\ttag lines).
// Java still constructs MultiWordChunker.getInstance — empty maps after load is correct
// (not "skip stage"). Do not invent multiword entries.
var arabicMultiWordChunkerSettings = disambiguation.MultiWordChunkerSettings{
	AllowFirstCapitalized: false,
	AllowAllUppercase:     false,
	AllowTitlecase:        false,
	// DefaultTag empty: phrase\ttag lines from official multiwords.txt (none today)
}

var (
	arMultiWordChunkerOnce sync.Once
	arMultiWordChunkerInst *disambiguation.MultiWordChunker
)

// OpenArabicMultiWordChunker ports MultiWordChunker.getInstance for
// /ar/multiwords.txt with ArabicHybridDisambiguator constructor defaults.
// Does not set remove-previous-tags or ignore-spelling.
func OpenArabicMultiWordChunker(r io.Reader) (*disambiguation.MultiWordChunker, error) {
	return disambiguation.NewMultiWordChunkerFromReader(r, arabicMultiWordChunkerSettings)
}

// LoadArabicMultiWordChunkerFromPath opens the official multiwords file
// and builds MultiWordChunker with ArabicHybridDisambiguator multiwords defaults.
func LoadArabicMultiWordChunkerFromPath(path string) (*disambiguation.MultiWordChunker, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return OpenArabicMultiWordChunker(f)
}

// ArabicMultiWordChunker returns the process-cached MultiWordChunker for
// official /ar/multiwords.txt (Java ArabicHybridDisambiguator.chunker field).
// Nil if the official resource is not discoverable.
// Official file may be comment-only → non-nil chunker with empty phrase maps.
func ArabicMultiWordChunker() *disambiguation.MultiWordChunker {
	arMultiWordChunkerOnce.Do(func() {
		p := DiscoverArabicMultiwords()
		if p == "" {
			return
		}
		c, err := LoadArabicMultiWordChunkerFromPath(p)
		if err != nil || c == nil {
			return
		}
		arMultiWordChunkerInst = c
	})
	return arMultiWordChunkerInst
}

// DiscoverArabicMultiwords finds official ar/multiwords.txt
// (Java resource /ar/multiwords.txt used by ArabicHybridDisambiguator.chunker).
func DiscoverArabicMultiwords() string {
	if p := os.Getenv("LANG_AR_MULTIWORDS_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "ar",
		"src", "main", "resources", "org", "languagetool", "resource", "ar", "multiwords.txt")
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
