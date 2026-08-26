package de

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/chunking"
)

// WireGermanChunker installs German.createDefaultPostDisambiguationChunker()
// (GermanChunker) on lt.PostDisambiguationChunker — not on pre-disambig Chunker
// (Java getChunker() is null for German).
//
// GermanChunker runs full Java REGEXES1 + REGEXES2 via OpenRegex.
func WireGermanChunker(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	// Java: only createDefaultPostDisambiguationChunker returns GermanChunker.
	lt.PostDisambiguationChunker = chunking.NewGermanChunker()
}
