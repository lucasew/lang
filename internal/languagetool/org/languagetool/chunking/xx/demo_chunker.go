package xx

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/lucasew/lang/internal/languagetool/org/languagetool/chunking"
)

// DemoChunker ports org.languagetool.chunking.xx.DemoChunker —
// assigns chunk B-NP-singular to the word "chunkbar".
type DemoChunker struct{}

func NewDemoChunker() *DemoChunker { return &DemoChunker{} }

func (DemoChunker) AddChunkTags(tokenReadings []*languagetool.AnalyzedTokenReadings) {
	for _, tr := range tokenReadings {
		if tr != nil && tr.GetToken() == "chunkbar" {
			tr.SetChunkTags([]string{"B-NP-singular"})
		}
	}
}

var _ chunking.Chunker = DemoChunker{}
