package chunking

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

// Twin of org.languagetool.chunking.ChunkTag (+ ChunkTaggedToken toString).

func TestChunkTag(t *testing.T) {
	ct := NewChunkTag("B-NP")
	require.Equal(t, "B-NP", ct.GetChunkTag())
	require.False(t, ct.IsRegexpTag())
	require.Equal(t, "B-NP", ct.String())

	re := NewChunkTagRegexp("B-NP", true)
	require.True(t, re.IsRegexpTag())
	require.Equal(t, "B-NP[regex]", re.String())
	// equals ignores isRegexp — only chunkTag string
	require.True(t, ct.Equal(re))
	require.True(t, ct.Equal(NewChunkTag("B-NP")))
	require.False(t, ct.Equal(NewChunkTag("I-NP")))

	require.Panics(t, func() { NewChunkTag("") })
	require.Panics(t, func() { NewChunkTag("   ") }) // trim().isEmpty()

	tok := NewChunkTaggedToken("cats", []ChunkTag{ct}, nil)
	require.Equal(t, "cats", tok.GetToken())
	require.Nil(t, tok.GetReadings())
	require.Equal(t, "cats/B-NP", tok.String())

	tok2 := NewChunkTaggedToken("dogs", []ChunkTag{ct, NewChunkTag("NP")}, nil)
	require.Equal(t, "dogs/B-NP,NP", tok2.String())
}

func TestTokenPredicate(t *testing.T) {
	nn := "NNS"
	readings := languagetool.NewAnalyzedTokenReadings(languagetool.NewAnalyzedToken("cats", &nn, nil))
	token := NewChunkTaggedToken("cats", []ChunkTag{NewChunkTag("B-NP")}, readings)

	require.True(t, NewTokenPredicate("cats", true).Apply(token))
	require.True(t, NewTokenPredicate("string=CATS", false).Apply(token))
	require.True(t, NewTokenPredicate("chunk=B-NP", true).Apply(token))
	require.True(t, NewTokenPredicate("pos=NN", true).Apply(token))
	require.False(t, NewTokenPredicate("pos=VB", true).Apply(token))
	require.True(t, NewTokenPredicate("regex=c.*", false).Apply(token))
}
