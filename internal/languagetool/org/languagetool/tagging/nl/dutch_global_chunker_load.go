package nl

import (
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// Dutch GlobalChunker settings match Java DutchHybridDisambiguator:
//
//	MultiWordChunker.getInstance("/spelling_global.txt", false, true, false,
//	  MultiWordChunker.tagForNotAddingTags);
//	chunkerGlobal.setIgnoreSpelling(true);
//
// Differs from Dutch multiwords: allowFirstCapitalized is false (not true).
// Same settings as DE MultitokenGlobal (GermanRuleDisambiguator multitokenSpeller3).
var dutchGlobalChunkerSettings = disambiguation.MultiWordChunkerSettings{
	AllowFirstCapitalized: false,
	AllowAllUppercase:     true,
	AllowTitlecase:        false,
	DefaultTag:            disambiguation.TagForNotAddingTags,
}

var (
	nlGlobalChunkerOnce sync.Once
	nlGlobalChunkerInst *disambiguation.MultiWordChunker
)

// OpenDutchGlobalChunker ports MultiWordChunker.getInstance for
// /spelling_global.txt with DutchHybridDisambiguator constructor settings
// (allowFirstCapitalized=false) and setIgnoreSpelling(true).
func OpenDutchGlobalChunker(r io.Reader) (*disambiguation.MultiWordChunker, error) {
	c, err := disambiguation.NewMultiWordChunkerFromReader(r, dutchGlobalChunkerSettings)
	if err != nil {
		return nil, err
	}
	c.SetIgnoreSpelling(true)
	return c, nil
}

// LoadDutchGlobalChunkerFromPath opens the official spelling_global file
// and builds MultiWordChunker with DutchHybridDisambiguator GlobalChunker defaults.
func LoadDutchGlobalChunkerFromPath(path string) (*disambiguation.MultiWordChunker, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return OpenDutchGlobalChunker(f)
}

// DutchGlobalChunker returns the process-cached MultiWordChunker for
// official /spelling_global.txt (Java DutchHybridDisambiguator.chunkerGlobal field).
// Nil if the official resource is not discoverable.
func DutchGlobalChunker() *disambiguation.MultiWordChunker {
	nlGlobalChunkerOnce.Do(func() {
		p := DiscoverDutchGlobalChunker()
		if p == "" {
			return
		}
		c, err := LoadDutchGlobalChunkerFromPath(p)
		if err != nil || c == nil {
			return
		}
		nlGlobalChunkerInst = c
	})
	return nlGlobalChunkerInst
}

// DiscoverDutchGlobalChunker finds official spelling_global.txt
// (Java resource /spelling_global.txt used by DutchHybridDisambiguator.chunkerGlobal).
func DiscoverDutchGlobalChunker() string {
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
