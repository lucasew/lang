package en

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/chunking"
)

// RegisterEnglishChunker installs English.createDefaultChunker() twin on lt.Chunker.
// OpenNLP ChunkerME when third_party/opennlp-models/en-chunker.bin is present; else POS→BIO + EnglishChunkFilter (no invent).
func RegisterEnglishChunker(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	lt.Chunker = chunking.NewEnglishChunker()
}
