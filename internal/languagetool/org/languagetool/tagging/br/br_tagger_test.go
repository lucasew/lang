package br

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBretonTagger(t *testing.T) {
	r := NewBretonTagger(nil)
	require.Equal(t, BretonTaggerDictPath, r.GetDictionaryPath())
	require.Empty(t, r.TagWord("x"))
}
