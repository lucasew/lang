package zh

// Twin of ChineseTaggerTest — parses HanLP-encoded surface/pos tokens like Java.
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestChineseTagger_Tagger(t *testing.T) {
	// Java path: tokenizer emits "surface/pos"; tagger splits.
	got := NewChineseTagger().Tag([]string{"你好/i", "世界/n", "xyz"})
	require.Len(t, got, 3)
	require.Equal(t, "你好", got[0].GetToken())
	require.NotNil(t, got[0].GetReadings()[0].GetPOSTag())
	require.Equal(t, "i", *got[0].GetReadings()[0].GetPOSTag())
	require.Equal(t, "世界", got[1].GetToken())
	require.Equal(t, "n", *got[1].GetReadings()[0].GetPOSTag())
	// malformed (no slash) → space token like Java
	require.Equal(t, " ", got[2].GetToken())
	require.Equal(t, ChineseDictPath, NewChineseTagger().GetDictionaryPath())
}
