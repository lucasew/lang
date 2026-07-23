package gl

import (
	"io"
	"os"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// OpenGalicianMultiWordChunker ports MultiWordChunker.getInstance("/gl/multiwords.txt")
// as used by GalicianHybridDisambiguator:
// allowFirstCapitalized=false, allowAllUppercase=false, allowTitlecase=false.
func OpenGalicianMultiWordChunker(r io.Reader) (*disambiguation.MultiWordChunker, error) {
	return disambiguation.NewMultiWordChunkerFromReader(r, disambiguation.MultiWordChunkerSettings{
		AllowFirstCapitalized: false,
		AllowAllUppercase:     false,
		AllowTitlecase:        false,
	})
}

// LoadGalicianMultiWordChunkerFromPath opens the official multiwords file at path
// and builds MultiWordChunker with GalicianHybridDisambiguator defaults.
func LoadGalicianMultiWordChunkerFromPath(path string) (*disambiguation.MultiWordChunker, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return OpenGalicianMultiWordChunker(f)
}
