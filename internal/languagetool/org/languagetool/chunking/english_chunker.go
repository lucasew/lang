package chunking

import (
	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// EnglishChunker ports org.languagetool.chunking.EnglishChunker surface.
// Full OpenNLP chunker is deferred; assigns B-NP to sequences of tagged nouns/adjectives
// when AssignBasicNP is true, then applies EnglishChunkFilter.
type EnglishChunker struct {
	Filter *EnglishChunkFilter
	// AssignBasicNP enables simple POS-based NP assignment when no chunks present.
	AssignBasicNP bool
	// IsNounish reports whether a POS tag starts an NP (default: NN*).
	IsNounish func(posTag string) bool
}

func NewEnglishChunker() *EnglishChunker {
	return &EnglishChunker{
		Filter:        NewEnglishChunkFilter(),
		AssignBasicNP: true,
		IsNounish: func(pos string) bool {
			return len(pos) >= 2 && pos[0] == 'N' && pos[1] == 'N'
		},
	}
}

// AddChunkTags implements Chunker.
func (c *EnglishChunker) AddChunkTags(tokens []*languagetool.AnalyzedTokenReadings) {
	if c == nil || len(tokens) == 0 {
		return
	}
	// convert to ChunkTaggedToken
	var tagged []ChunkTaggedToken
	for _, t := range tokens {
		if t == nil {
			continue
		}
		var tags []ChunkTag
		for _, ct := range t.GetChunkTags() {
			tags = append(tags, NewChunkTag(ct))
		}
		tagged = append(tagged, NewChunkTaggedToken(t.GetToken(), tags, t))
	}
	if c.AssignBasicNP {
		tagged = c.assignBasicNP(tagged)
	}
	if c.Filter != nil {
		tagged = c.Filter.Filter(tagged)
	}
	// write back chunk tags
	for i, t := range tagged {
		if i >= len(tokens) || tokens[i] == nil {
			continue
		}
		var strs []string
		for _, ct := range t.ChunkTags {
			strs = append(strs, ct.GetChunkTag())
		}
		tokens[i].SetChunkTags(strs)
	}
}

func (c *EnglishChunker) assignBasicNP(tokens []ChunkTaggedToken) []ChunkTaggedToken {
	if c.IsNounish == nil {
		return tokens
	}
	out := make([]ChunkTaggedToken, len(tokens))
	copy(out, tokens)
	inNP := false
	for i, t := range out {
		pos := ""
		if t.Readings != nil {
			for _, r := range t.Readings.GetReadings() {
				if r != nil && r.GetPOSTag() != nil {
					pos = *r.GetPOSTag()
					break
				}
			}
		}
		if c.IsNounish(pos) {
			if !inNP {
				out[i].ChunkTags = []ChunkTag{NewChunkTag("B-NP")}
				inNP = true
			} else {
				out[i].ChunkTags = []ChunkTag{NewChunkTag("I-NP")}
			}
		} else {
			inNP = false
		}
	}
	return out
}

var _ Chunker = (*EnglishChunker)(nil)
