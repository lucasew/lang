package en

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/chunking"
)

// RegisterEnglishChunker installs Java English.createDefaultChunker() soft port
// (POS-driven OpenNLP-like BIO + EnglishChunkFilter) on lt.Chunker.
func RegisterEnglishChunker(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	lt.Chunker = chunking.NewEnglishChunker()
}
