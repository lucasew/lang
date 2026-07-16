package chunking

import (
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestChunkTag_AndToken(t *testing.T) {
	ct := NewChunkTag("B-NP")
	require.Equal(t, "B-NP", ct.GetChunkTag())
	require.Equal(t, "B-NP", ct.String())
	require.Equal(t, "B-NP[regex]", NewChunkTagRegexp("B-NP", true).String())
	require.Panics(t, func() { NewChunkTag("") })

	tok := NewChunkTaggedToken("cats", []ChunkTag{ct}, nil)
	require.Equal(t, "cats/B-NP", tok.String())
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
