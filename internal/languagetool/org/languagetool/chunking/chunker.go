package chunking

import "github.com/lucasew/lang/internal/languagetool/org/languagetool"

// Chunker ports org.languagetool.chunking.Chunker.
type Chunker interface {
	AddChunkTags(sentenceTokenReadings []*languagetool.AnalyzedTokenReadings)
}

// FuncChunker adapts a function to Chunker.
type FuncChunker func(sentenceTokenReadings []*languagetool.AnalyzedTokenReadings)

func (f FuncChunker) AddChunkTags(sentenceTokenReadings []*languagetool.AnalyzedTokenReadings) {
	if f != nil {
		f(sentenceTokenReadings)
	}
}
