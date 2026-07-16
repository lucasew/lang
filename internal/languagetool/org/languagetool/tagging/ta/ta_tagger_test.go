package ta

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTamilTagger(t *testing.T) {
	r := NewTamilTagger(nil)
	require.Equal(t, TamilTaggerDictPath, r.GetDictionaryPath())
}
