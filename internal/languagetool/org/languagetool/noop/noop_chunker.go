package noop

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/chunking"
)

// NoopChunker ports org.languagetool.noop.NoopChunker.
type NoopChunker struct{}

func NewNoopChunker() *NoopChunker { return &NoopChunker{} }

func (NoopChunker) AddChunkTags(_ []*languagetool.AnalyzedTokenReadings) {}

var _ chunking.Chunker = NoopChunker{}
