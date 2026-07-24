package chunking

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestEnglishChunkFilterSingular(t *testing.T) {
	pos := "NN"
	tok := languagetool.NewAnalyzedToken("dog", &pos, nil)
	rd := languagetool.NewAnalyzedTokenReadingsAt(tok, 0)
	tokens := []ChunkTaggedToken{
		NewChunkTaggedToken("dog", []ChunkTag{NewChunkTag("B-NP")}, rd),
	}
	out := NewEnglishChunkFilter().Filter(tokens)
	require.Len(t, out, 1)
	require.Equal(t, "B-NP-singular", out[0].ChunkTags[0].GetChunkTag())
}

func TestEnglishChunkFilterPlural(t *testing.T) {
	pos := "NNS"
	tok := languagetool.NewAnalyzedToken("dogs", &pos, nil)
	rd := languagetool.NewAnalyzedTokenReadingsAt(tok, 0)
	tokens := []ChunkTaggedToken{
		NewChunkTaggedToken("dogs", []ChunkTag{NewChunkTag("B-NP")}, rd),
	}
	out := NewEnglishChunkFilter().Filter(tokens)
	require.Equal(t, "B-NP-plural", out[0].ChunkTags[0].GetChunkTag())
}
