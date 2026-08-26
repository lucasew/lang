package crh

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCrimeanTatarTagger(t *testing.T) {
	r := NewCrimeanTatarTagger(nil)
	require.Equal(t, CrimeanTatarTaggerDictPath, r.GetDictionaryPath())
	require.Empty(t, r.TagWord("x"))
}
