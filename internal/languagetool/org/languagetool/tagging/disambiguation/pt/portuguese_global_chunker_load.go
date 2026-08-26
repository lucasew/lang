package pt

import (
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// Portuguese GlobalChunker settings match Java PortugueseHybridDisambiguator:
//
//	MultiWordChunker.getInstance("/spelling_global.txt", false, true, true, "NPCN000");
//	chunkerGlobal.setIgnoreSpelling(true);
//	// constructor does NOT call setRemovePreviousTags on chunkerGlobal
//
// Differs from Spanish GlobalChunker: allowTitlecase=true; SetIgnoreSpelling(true).
// Differs from Dutch GlobalChunker: DefaultTag "NPCN000" (not tagForNotAddingTags).
// Note: allowTitlecase only affects variants when allowFirstCapitalized is also true
// (Java MultiWordChunker.getTokenLettercaseVariants nests titlecase under first-cap);
// with allowFirstCapitalized=false, titlecase flag is stored but generates no variants.
var portugueseGlobalChunkerSettings = disambiguation.MultiWordChunkerSettings{
	AllowFirstCapitalized: false,
	AllowAllUppercase:     true,
	AllowTitlecase:        true,
	DefaultTag:            "NPCN000",
}

var (
	ptGlobalChunkerOnce sync.Once
	ptGlobalChunkerInst *disambiguation.MultiWordChunker
)

// OpenPortugueseGlobalChunker ports MultiWordChunker.getInstance for
// /spelling_global.txt with PortugueseHybridDisambiguator constructor settings
// (allowFirstCapitalized=false, allowAllUppercase=true, allowTitlecase=true,
// DefaultTag "NPCN000") and setIgnoreSpelling(true).
func OpenPortugueseGlobalChunker(r io.Reader) (*disambiguation.MultiWordChunker, error) {
	c, err := disambiguation.NewMultiWordChunkerFromReader(r, portugueseGlobalChunkerSettings)
	if err != nil {
		return nil, err
	}
	c.SetIgnoreSpelling(true)
	return c, nil
}

// LoadPortugueseGlobalChunkerFromPath opens the official spelling_global file
// and builds MultiWordChunker with PortugueseHybridDisambiguator GlobalChunker defaults.
func LoadPortugueseGlobalChunkerFromPath(path string) (*disambiguation.MultiWordChunker, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return OpenPortugueseGlobalChunker(f)
}

// PortugueseGlobalChunker returns the process-cached MultiWordChunker for
// official /spelling_global.txt (Java PortugueseHybridDisambiguator.chunkerGlobal field).
// Nil if the official resource is not discoverable.
func PortugueseGlobalChunker() *disambiguation.MultiWordChunker {
	ptGlobalChunkerOnce.Do(func() {
		p := DiscoverPortugueseGlobalChunker()
		if p == "" {
			return
		}
		c, err := LoadPortugueseGlobalChunkerFromPath(p)
		if err != nil || c == nil {
			return
		}
		ptGlobalChunkerInst = c
	})
	return ptGlobalChunkerInst
}

// DiscoverPortugueseGlobalChunker finds official spelling_global.txt
// (Java resource /spelling_global.txt used by PortugueseHybridDisambiguator.chunkerGlobal).
func DiscoverPortugueseGlobalChunker() string {
	if p := os.Getenv("LANG_SPELLING_GLOBAL_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-core",
		"src", "main", "resources", "org", "languagetool", "resource", "spelling_global.txt")
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
