package ru

import (
	"io"
	"os"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"
)

// OpenRussianMultiWordChunker ports MultiWordChunker.getInstance("/ru/multiwords.txt")
// as used by RussianHybridDisambiguator:
// allowFirstCapitalized=false, allowAllUppercase=false, allowTitlecase=false.
func OpenRussianMultiWordChunker(r io.Reader) (*disambiguation.MultiWordChunker, error) {
	return disambiguation.NewMultiWordChunkerFromReader(r, disambiguation.MultiWordChunkerSettings{
		AllowFirstCapitalized: false,
		AllowAllUppercase:     false,
		AllowTitlecase:        false,
	})
}

// LoadRussianMultiWordChunkerFromPath opens the official multiwords file at path
// and builds MultiWordChunker with RussianHybridDisambiguator defaults.
func LoadRussianMultiWordChunkerFromPath(path string) (*disambiguation.MultiWordChunker, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return OpenRussianMultiWordChunker(f)
}
