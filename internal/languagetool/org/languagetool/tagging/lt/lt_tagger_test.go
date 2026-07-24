package lt

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLithuanianTagger(t *testing.T) {
	r := NewLithuanianTagger(nil)
	require.Equal(t, LithuanianTaggerDictPath, r.GetDictionaryPath())
}
