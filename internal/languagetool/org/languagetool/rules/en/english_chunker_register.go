package en

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/chunking"
)

// RegisterEnglishChunker installs English.createDefaultChunker() twin on lt.Chunker.
// OpenNLP maxent is not run yet; POS→BIO + EnglishChunkFilter only (incomplete, no soft invent).
func RegisterEnglishChunker(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	lt.Chunker = chunking.NewEnglishChunker()
}
