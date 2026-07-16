package chunking

import "strings"

// ChunkTag ports org.languagetool.chunking.ChunkTag.
type ChunkTag struct {
	ChunkTag string
	IsRegexp bool
}

func NewChunkTag(chunkTag string) ChunkTag {
	return NewChunkTagRegexp(chunkTag, false)
}

func NewChunkTagRegexp(chunkTag string, isRegexp bool) ChunkTag {
	if strings.TrimSpace(chunkTag) == "" {
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
