package pl

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpenPolishSynthesizerFromDir_Missing(t *testing.T) {
	require.Nil(t, OpenPolishSynthesizerFromDir(""))
	require.Nil(t, OpenPolishSynthesizerFromDir(t.TempDir()))
	require.Nil(t, OpenPolishSynthesizerFromDictPath(""))
}

func TestOpenPolishSynthesizer_FromDict(t *testing.T) {
	p := DiscoverPolishSynthDict()
	if p == "" {
		t.Skip("polish_synth.dict not discoverable")
	}
	s := OpenPolishSynthesizerFromDictPath(p)
	require.NotNil(t, s)
	require.NotNil(t, s.Lookup)
	require.NotEmpty(t, s.PossibleTags)
	require.Equal(t, ResourceFilename, s.ResourceFileName)
	require.Equal(t, TagsFileName, s.TagFileName)
	// getPosTagCorrection retained on Polish type
	require.Contains(t, s.GetPosTagCorrection("foo.bar"), ".*|.*")
}

func TestPolishSynthesizer_Instance(t *testing.T) {
	s := Instance()
	if s == nil {
		t.Skip("polish_synth.dict / official PL synth resources not available")
	}
	require.Same(t, s, INSTANCE)
	require.NotNil(t, s.Lookup)
	require.NotEmpty(t, s.PossibleTags)
	require.Equal(t, ResourceFilename, s.ResourceFileName)
	require.Equal(t, TagsFileName, s.TagFileName)
}
