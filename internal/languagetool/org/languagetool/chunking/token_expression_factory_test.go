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
	// Java LogicExpression OR
	require.True(t, f.Create("string=the|string=a").Apply(tok))
	require.True(t, f.Create("the|a").Apply(tok))
	require.True(t, f.Create("the").Apply(tok))
	// Java AND is & not space
	require.False(t, f.Create("string=the & string=a").Apply(tok))
	require.True(t, f.Create("string=the & chunk=B-NP").Apply(tok))
	require.True(t, f.Create("chunk=B-NP & !pos=VB").Apply(tok)) // no POS → pos fails → !pos true? pos check false so !pos true
}

func TestLogicExpression_AndOrNot(t *testing.T) {
	f := NewTokenExpressionFactory(false)
	tok := NewChunkTaggedToken("Hund", []ChunkTag{NewChunkTag("B-NP")}, nil)
	// surface + chunk
	require.True(t, f.Create("hund & chunk=B-NP").Apply(tok))
	require.False(t, f.Create("katze | maus").Apply(tok))
	require.True(t, f.Create("!katze").Apply(tok))
	require.True(t, f.Create("hund | katze").Apply(tok))
}

func TestOpenRegex_Basic(t *testing.T) {
	factory := NewChunkTokenFactory(false)
	// <det> <noun>+
	re := CompileOpenRegex("<the> <dog|cat>+", factory)
	tokens := []ChunkTaggedToken{
		NewChunkTaggedToken("the", nil, nil),
		NewChunkTaggedToken("dog", nil, nil),
		NewChunkTaggedToken("cat", nil, nil),
		NewChunkTaggedToken("runs", nil, nil),
	}
	ms := re.FindAll(tokens)
	require.Len(t, ms, 1)
	require.Equal(t, 0, ms[0].Start)
	require.Equal(t, 3, ms[0].End)

	// German-style NP expansion
	re2 := CompileOpenRegex(ExpandGermanChunkSyntax("<NP>"), factory)
	npToks := []ChunkTaggedToken{
		NewChunkTaggedToken("Ein", []ChunkTag{NewChunkTag("B-NP")}, nil),
		NewChunkTaggedToken("Haus", []ChunkTag{NewChunkTag("I-NP")}, nil),
		NewChunkTaggedToken("steht", []ChunkTag{NewChunkTag("O")}, nil),
	}
	ms2 := re2.FindAll(npToks)
	require.Len(t, ms2, 1)
	require.Equal(t, SeqMatch{0, 2}, ms2[0])
}
