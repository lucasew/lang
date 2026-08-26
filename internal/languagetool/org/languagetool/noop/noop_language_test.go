package noop

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNoopLanguage(t *testing.T) {
	n := NewNoopLanguage()
	require.Equal(t, ShortCode, n.GetShortCode())
	require.Equal(t, "NoopLanguage", n.GetName())
	require.Empty(t, n.CreateDefaultWordTokenizer().Tokenize("hello"))
	require.Equal(t, []string{"hello"}, n.CreateDefaultSentenceTokenizer().Tokenize("hello"))
	sent := n.CreateDefaultDisambiguator().Disambiguate(nil)
	require.Nil(t, sent)
	tags, err := n.CreateDefaultTagger().Tag([]string{"a", "b"})
	require.NoError(t, err)
	require.Len(t, tags, 2)
	require.Nil(t, tags[0].GetReadings()[0].GetPOSTag())
}
