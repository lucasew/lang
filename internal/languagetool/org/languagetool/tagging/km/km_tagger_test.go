package km

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKhmerTagger(t *testing.T) {
	r := NewKhmerTagger(nil)
	require.Equal(t, KhmerTaggerDictPath, r.GetDictionaryPath())
	require.Empty(t, r.TagWord("x"))
}
