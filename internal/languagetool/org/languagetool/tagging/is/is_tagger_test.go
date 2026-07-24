package is

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIcelandicTagger(t *testing.T) {
	r := NewIcelandicTagger(nil)
	require.Equal(t, IcelandicTaggerDictPath, r.GetDictionaryPath())
}
