package chunking

import (
	"strings"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
)

// ChunkTaggedToken ports org.languagetool.chunking.ChunkTaggedToken.
type ChunkTaggedToken struct {
	Token     string
	ChunkTags []ChunkTag
	// Readings may be nil when tokenization does not map 1:1.
	Readings *languagetool.AnalyzedTokenReadings
}

func NewChunkTaggedToken(token string, chunkTags []ChunkTag, readings *languagetool.AnalyzedTokenReadings) ChunkTaggedToken {
	return ChunkTaggedToken{
		Token:     token,
		ChunkTags: append([]ChunkTag(nil), chunkTags...),
		Readings:  readings,
	}
}

func (c ChunkTaggedToken) GetToken() string { return c.Token }
func (c ChunkTaggedToken) GetChunkTags() []ChunkTag {
	return c.ChunkTags
}
func (c ChunkTaggedToken) GetReadings() *languagetool.AnalyzedTokenReadings {
	return c.Readings
}

func (c ChunkTaggedToken) String() string {
	parts := make([]string, len(c.ChunkTags))
	for i, t := range c.ChunkTags {
		parts[i] = t.String()
	}
	return c.Token + "/" + strings.Join(parts, ",")
}
