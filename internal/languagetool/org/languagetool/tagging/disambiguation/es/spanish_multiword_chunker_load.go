package es

import (
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// Spanish multiwords settings match Java SpanishHybridDisambiguator:
//
//	MultiWordChunker.getInstance("/es/multiwords.txt", true, true, false);
//	// default tag: null (normal multiword open-close tags — not tagForNotAddingTags)
//	chunker.setRemovePreviousTags(true);
//	// NO setIgnoreSpelling (unlike NL)
var spanishMultiWordChunkerSettings = disambiguation.MultiWordChunkerSettings{
	AllowFirstCapitalized: true,
	AllowAllUppercase:     true,
	AllowTitlecase:        false,
	// DefaultTag empty: phrase;tag / phrase\ttag lines from official multiwords.txt
}

var (
	esMultiWordChunkerOnce sync.Once
	esMultiWordChunkerInst *disambiguation.MultiWordChunker
)

// OpenSpanishMultiWordChunker ports MultiWordChunker.getInstance for
// /es/multiwords.txt with SpanishHybridDisambiguator constructor settings
// and setRemovePreviousTags(true). Does not set ignore-spelling.
func OpenSpanishMultiWordChunker(r io.Reader) (*disambiguation.MultiWordChunker, error) {
	c, err := disambiguation.NewMultiWordChunkerFromReader(r, spanishMultiWordChunkerSettings)
	if err != nil {
		return nil, err
	}
	c.SetRemovePreviousTags(true)
	return c, nil
}

// LoadSpanishMultiWordChunkerFromPath opens the official multiwords file
// and builds MultiWordChunker with SpanishHybridDisambiguator multiwords defaults.
func LoadSpanishMultiWordChunkerFromPath(path string) (*disambiguation.MultiWordChunker, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return OpenSpanishMultiWordChunker(f)
}

// SpanishMultiWordChunker returns the process-cached MultiWordChunker for
// official /es/multiwords.txt (Java SpanishHybridDisambiguator.chunker field).
// Nil if the official resource is not discoverable.
func SpanishMultiWordChunker() *disambiguation.MultiWordChunker {
	esMultiWordChunkerOnce.Do(func() {
		p := DiscoverSpanishMultiwords()
		if p == "" {
			return
		}
		c, err := LoadSpanishMultiWordChunkerFromPath(p)
		if err != nil || c == nil {
			return
		}
		esMultiWordChunkerInst = c
	})
	return esMultiWordChunkerInst
}

// DiscoverSpanishMultiwords finds official es/multiwords.txt
// (Java resource /es/multiwords.txt used by SpanishHybridDisambiguator.chunker).
func DiscoverSpanishMultiwords() string {
	if p := os.Getenv("LANG_ES_MULTIWORDS_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "es",
		"src", "main", "resources", "org", "languagetool", "resource", "es", "multiwords.txt")
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
