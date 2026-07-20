package en

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/chunking"
)

// RegisterEnglishChunker installs English.createDefaultChunker() twin on lt.Chunker.
// Full OpenNLP path (en-token + en-pos-maxent + en-chunker) when models present;
// else POS→BIO + EnglishChunkFilter (no invent).
func RegisterEnglishChunker(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	lt.Chunker = chunking.NewEnglishChunker()
}
