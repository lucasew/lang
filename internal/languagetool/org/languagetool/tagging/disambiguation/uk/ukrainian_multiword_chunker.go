package uk

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/tagging/disambiguation"

// UkrainianMultiwordChunker wraps MultiWordChunker for /uk/multiwords.txt style data.
type UkrainianMultiwordChunker = disambiguation.MultiWordChunker

func NewUkrainianMultiwordChunker(lines []string) *disambiguation.MultiWordChunker {
	return disambiguation.NewMultiWordChunker(lines, disambiguation.MultiWordChunkerSettings{
		AllowFirstCapitalized: true,
	})
}
