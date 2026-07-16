package chunking

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// RussianChunker ports org.languagetool.chunking.RussianChunker (simplified OpenNLP-free).
// Assigns basic NP/VP chunks from Russian POS prefixes.
type RussianChunker struct{}

func NewRussianChunker() *RussianChunker { return &RussianChunker{} }

func (c *RussianChunker) AddChunkTags(tokens []*languagetool.AnalyzedTokenReadings) {
	if c == nil || len(tokens) == 0 {
		return
	}
	inNP, inVP := false, false
	for _, t := range tokens {
		if t == nil {
			continue
		}
		pos := firstPOS(t)
		switch {
		case isRussianNoun(pos) || isRussianAdj(pos):
			if !inNP {
				t.SetChunkTags([]string{"B-NP"})
				inNP = true
			} else {
				t.SetChunkTags([]string{"I-NP"})
			}
			inVP = false
		case isRussianVerb(pos):
			if !inVP {
				t.SetChunkTags([]string{"B-VP"})
				inVP = true
			} else {
				t.SetChunkTags([]string{"I-VP"})
			}
			inNP = false
		default:
			inNP, inVP = false, false
		}
	}
}

func isRussianNoun(pos string) bool {
	return strings.Contains(pos, "NN") || strings.HasPrefix(pos, "S") // S for noun in some tagsets
}

func isRussianAdj(pos string) bool {
	return strings.Contains(pos, "ADJ") || strings.HasPrefix(pos, "A")
}

func isRussianVerb(pos string) bool {
	return strings.Contains(pos, "V") && !strings.Contains(pos, "ADV")
}

var _ Chunker = (*RussianChunker)(nil)
