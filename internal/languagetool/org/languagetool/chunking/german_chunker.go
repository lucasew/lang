package chunking

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// GermanChunker ports org.languagetool.chunking.GermanChunker (simplified OpenNLP-free).
// Assigns B-NP/I-NP to German noun phrases (SUB/EIG/ART/ADJ sequences).
type GermanChunker struct {
	// IsNPStart reports POS tags that open an NP.
	IsNPStart func(pos string) bool
	// IsNPCont reports POS tags that continue an NP.
	IsNPCont func(pos string) bool
}

func NewGermanChunker() *GermanChunker {
	return &GermanChunker{
		IsNPStart: func(pos string) bool {
			return strings.HasPrefix(pos, "ART") || strings.HasPrefix(pos, "SUB") ||
				strings.HasPrefix(pos, "EIG") || strings.HasPrefix(pos, "PRO:PER")
		},
		IsNPCont: func(pos string) bool {
			return strings.HasPrefix(pos, "SUB") || strings.HasPrefix(pos, "EIG") ||
				strings.HasPrefix(pos, "ADJ") || strings.HasPrefix(pos, "PA2") ||
				strings.HasPrefix(pos, "ART") || strings.HasPrefix(pos, "PRO")
		},
	}
}

func (c *GermanChunker) AddChunkTags(tokens []*languagetool.AnalyzedTokenReadings) {
	if c == nil || len(tokens) == 0 {
		return
	}
	inNP := false
	for _, t := range tokens {
		if t == nil {
			continue
		}
		pos := firstPOS(t)
		if !inNP && c.IsNPStart != nil && c.IsNPStart(pos) {
			t.SetChunkTags([]string{"B-NP"})
			inNP = true
			continue
		}
		if inNP && c.IsNPCont != nil && c.IsNPCont(pos) {
			t.SetChunkTags([]string{"I-NP"})
			continue
		}
		inNP = false
	}
}

func firstPOS(t *languagetool.AnalyzedTokenReadings) string {
	for _, r := range t.GetReadings() {
		if r != nil && r.GetPOSTag() != nil {
			return *r.GetPOSTag()
		}
	}
	return ""
}

var _ Chunker = (*GermanChunker)(nil)
