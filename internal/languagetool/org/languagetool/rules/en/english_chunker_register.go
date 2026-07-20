package en

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/chunking"
)

// RegisterEnglishChunker installs English.createDefaultChunker() twin on lt.Chunker.
// Java OpenNLP path only (en-token + en-pos-maxent + en-chunker); missing models = incomplete.
func RegisterEnglishChunker(lt *languagetool.JLanguageTool) {
	if lt == nil {
		return
	}
	lt.Chunker = chunking.NewEnglishChunker()
}
