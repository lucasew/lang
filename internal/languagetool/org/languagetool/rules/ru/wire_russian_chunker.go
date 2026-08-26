package ru

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/chunking"
)

// WireRussianChunker installs Russian.createDefaultPostDisambiguationChunker()
// (RussianChunker) on lt.PostDisambiguationChunker — not on pre-disambig Chunker
// (Java getChunker() is null for Russian).
func WireRussianChunker(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	lt.PostDisambiguationChunker = chunking.NewRussianChunker()
}
