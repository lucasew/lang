package language

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDemoLanguageSurface(t *testing.T) {
	d := NewDemo()
	require.Equal(t, "xx", d.GetShortCode())
	require.Equal(t, "Testlanguage", d.GetName())
	require.Equal(t, []string{"XX"}, d.GetCountries())
	require.NotNil(t, d.CreateDefaultTagger())
	require.NotNil(t, d.CreateDefaultWordTokenizer())
	require.NotNil(t, d.CreateDefaultSentenceTokenizer())
	require.NotNil(t, d.CreateDefaultChunker())
	require.NotNil(t, d.CreateDefaultDisambiguator())
}
