package gl

import (
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// Galician multiwords settings match Java GalicianHybridDisambiguator:
//
//	MultiWordChunker.getInstance("/gl/multiwords.txt");
//	// defaults: allowFirstCapitalized=false, allowAllUppercase=false, allowTitlecase=false
//	// default tag: null (normal multiword open-close tags — not tagForNotAddingTags)
//	// NO setRemovePreviousTags, NO setIgnoreSpelling
var galicianMultiWordChunkerSettings = disambiguation.MultiWordChunkerSettings{
	AllowFirstCapitalized: false,
	AllowAllUppercase:     false,
	AllowTitlecase:        false,
	// DefaultTag empty: phrase\ttag lines from official multiwords.txt
}

var (
	glMultiWordChunkerOnce sync.Once
	glMultiWordChunkerInst *disambiguation.MultiWordChunker
)

// OpenGalicianMultiWordChunker ports MultiWordChunker.getInstance for
// /gl/multiwords.txt with GalicianHybridDisambiguator constructor defaults.
// Does not set remove-previous-tags or ignore-spelling.
func OpenGalicianMultiWordChunker(r io.Reader) (*disambiguation.MultiWordChunker, error) {
	return disambiguation.NewMultiWordChunkerFromReader(r, galicianMultiWordChunkerSettings)
}

// LoadGalicianMultiWordChunkerFromPath opens the official multiwords file
// and builds MultiWordChunker with GalicianHybridDisambiguator multiwords defaults.
func LoadGalicianMultiWordChunkerFromPath(path string) (*disambiguation.MultiWordChunker, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return OpenGalicianMultiWordChunker(f)
}

// GalicianMultiWordChunker returns the process-cached MultiWordChunker for
// official /gl/multiwords.txt (Java GalicianHybridDisambiguator.chunker field).
// Nil if the official resource is not discoverable.
func GalicianMultiWordChunker() *disambiguation.MultiWordChunker {
	glMultiWordChunkerOnce.Do(func() {
		p := DiscoverGalicianMultiwords()
		if p == "" {
			return
		}
		c, err := LoadGalicianMultiWordChunkerFromPath(p)
		if err != nil || c == nil {
			return
		}
		glMultiWordChunkerInst = c
	})
	return glMultiWordChunkerInst
}

// DiscoverGalicianMultiwords finds official gl/multiwords.txt
// (Java resource /gl/multiwords.txt used by GalicianHybridDisambiguator.chunker).
func DiscoverGalicianMultiwords() string {
	if p := os.Getenv("LANG_GL_MULTIWORDS_FILE"); p != "" {
		if st, err := os.Stat(p); err == nil && st.Mode().IsRegular() {
			return p
		}
	}
	rel := filepath.Join("inspiration", "languagetool", "languagetool-language-modules", "gl",
		"src", "main", "resources", "org", "languagetool", "resource", "gl", "multiwords.txt")
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
