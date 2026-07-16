package ml

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMalayalamTagger(t *testing.T) {
	r := NewMalayalamTagger(nil)
	require.Equal(t, MalayalamTaggerDictPath, r.GetDictionaryPath())
	require.Empty(t, r.TagWord("x"))
}
