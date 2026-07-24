package chunking

import "github.com/lucasew/lang/internal/languagetool/org/languagetool/tools"

// ChunkTag ports org.languagetool.chunking.ChunkTag.
type ChunkTag struct {
	ChunkTag string
	IsRegexp bool
}

func NewChunkTag(chunkTag string) ChunkTag {
	return NewChunkTagRegexp(chunkTag, false)
}

func NewChunkTagRegexp(chunkTag string, isRegexp bool) ChunkTag {
	// Java: chunkTag == null || chunkTag.trim().isEmpty()
	if tools.JavaStringTrimIsEmpty(chunkTag) {
		panic("chunkTag cannot be null or empty: '" + chunkTag + "'")
	}
	return ChunkTag{ChunkTag: chunkTag, IsRegexp: isRegexp}
}

func (c ChunkTag) GetChunkTag() string { return c.ChunkTag }
func (c ChunkTag) IsRegexpTag() bool   { return c.IsRegexp }

func (c ChunkTag) Equal(o ChunkTag) bool {
	return c.ChunkTag == o.ChunkTag
}

func (c ChunkTag) String() string {
	if c.IsRegexp {
		return c.ChunkTag + "[regex]"
	}
	return c.ChunkTag
}
