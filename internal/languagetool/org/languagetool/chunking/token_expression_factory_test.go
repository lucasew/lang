package chunking

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTokenExpressionFactory(t *testing.T) {
	f := NewTokenExpressionFactory(false)
	tok := NewChunkTaggedToken("The", []ChunkTag{NewChunkTag("B-NP")}, nil)

	require.True(t, f.Create("string=the").Apply(tok))
	require.False(t, f.Create("string=a").Apply(tok))
	require.True(t, f.Create("string=the|string=a").Apply(tok))
	require.True(t, f.Create("the").Apply(tok))
	require.False(t, f.Create("string=the string=a").Apply(tok))
	require.True(t, f.Create("string=the chunk=B-NP").Apply(tok))
}
