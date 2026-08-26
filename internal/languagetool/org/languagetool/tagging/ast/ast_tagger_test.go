package ast

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAsturianTagger(t *testing.T) {
	r := NewAsturianTagger(nil)
	require.Equal(t, AsturianTaggerDictPath, r.GetDictionaryPath())
	require.Empty(t, r.TagWord("x"))
}
