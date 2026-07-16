package tl

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTagalogTagger(t *testing.T) {
	r := NewTagalogTagger(nil)
	require.Equal(t, TagalogTaggerDictPath, r.GetDictionaryPath())
	require.Empty(t, r.TagWord("x"))
}
