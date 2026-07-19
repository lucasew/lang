package de

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/chunking"
)

// WireGermanChunker installs German.createDefaultPostDisambiguationChunker()
// (GermanChunker) on lt.PostDisambiguationChunker — not on pre-disambig Chunker
// (Java getChunker() is null for German).
//
// The Go GermanChunker implements Java REGEXES1 fully (as sequential matchers)
// plus a growing REGEXES2 subset (NPS/NPP/PP, genitive, late PP). Remaining
// OpenRegex-only paths stay incomplete (not invent). Always available.
func WireGermanChunker(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	// Java: only createDefaultPostDisambiguationChunker returns GermanChunker.
	lt.PostDisambiguationChunker = chunking.NewGermanChunker()
}
