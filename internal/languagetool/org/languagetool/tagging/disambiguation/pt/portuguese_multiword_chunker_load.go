package pt

import (
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// Portuguese multiwords settings match Java PortugueseHybridDisambiguator:
//
//	MultiWordChunker.getInstance("/pt/multiwords.txt", true, true, true);
//	// default tag: null (normal multiword open-close tags — not tagForNotAddingTags)
//	chunker.setRemovePreviousTags(true);
//	chunker.setIgnoreSpelling(true);
//
// Note allowTitlecase=true (unlike ES multiwords which is false).
var portugueseMultiWordChunkerSettings = disambiguation.MultiWordChunkerSettings{
	AllowFirstCapitalized: true,
	AllowAllUppercase:     true,
	AllowTitlecase:        true,
	// DefaultTag empty: phrase\ttag lines from official multiwords.txt
}

var (
	ptMultiWordChunkerOnce sync.Once
	ptMultiWordChunkerInst *disambiguation.MultiWordChunker
)

// OpenPortugueseMultiWordChunker ports MultiWordChunker.getInstance for
// /pt/multiwords.txt with PortugueseHybridDisambiguator constructor settings,
// setRemovePreviousTags(true), and setIgnoreSpelling(true).
func OpenPortugueseMultiWordChunker(r io.Reader) (*disambiguation.MultiWordChunker, error) {
	c, err := disambiguation.NewMultiWordChunkerFromReader(r, portugueseMultiWordChunkerSettings)
	if err != nil {
		return nil, err
	}
	c.SetRemovePreviousTags(true)
	c.SetIgnoreSpelling(true)
	return c, nil
}

// LoadPortugueseMultiWordChunkerFromPath opens the official multiwords file
// and builds MultiWordChunker with PortugueseHybridDisambiguator multiwords defaults.
func LoadPortugueseMultiWordChunkerFromPath(path string) (*disambiguation.MultiWordChunker, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return OpenPortugueseMultiWordChunker(f)
}

// PortugueseMultiWordChunker returns the process-cached MultiWordChunker for
// official /pt/multiwords.txt (Java PortugueseHybridDisambiguator.chunker field).
// Nil if the official resource is not discoverable.
func PortugueseMultiWordChunker() *disambiguation.MultiWordChunker {
	ptMultiWordChunkerOnce.Do(func() {
		p := DiscoverPortugueseMultiwords()
		if p == "" {
			return
		}
		c, err := LoadPortugueseMultiWordChunkerFromPath(p)
		if err != nil || c == nil {
			return
		}
		ptMultiWordChunkerInst = c
	})
	return ptMultiWordChunkerInst
}

// DiscoverPortugueseMultiwords finds official pt/multiwords.txt
// (Java resource /pt/multiwords.txt used by PortugueseHybridDisambiguator.chunker).
func DiscoverPortugueseMultiwords() string {
	if p := os.Getenv("LANG_PT_MULTIWORDS_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "pt",
		"src", "main", "resources", "org", "languagetool", "resource", "pt", "multiwords.txt")
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
