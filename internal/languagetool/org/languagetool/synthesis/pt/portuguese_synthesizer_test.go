package pt

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Constructor path constants (Java RESOURCE_FILENAME / TAGS_FILE_NAME / SOR_FILE_NAME) without dict.
func TestPortugueseSynthesizer_ResourcePaths(t *testing.T) {
	s := NewPortugueseSynthesizer(nil)
	require.Equal(t, PortugueseSynthDict, s.ResourceFileName)
	require.Equal(t, PortugueseTagsFile, s.TagFileName)
	require.Equal(t, PortugueseSorFile, s.SorFileName)
	require.Equal(t, PortugueseSorFile, s.GetSorFileName())
	require.Equal(t, "pt", s.LangShortCode)
}

// Open path fail-closed + successful load from discovered dict.
func TestOpenPortugueseSynthesizer_FromDict(t *testing.T) {
	require.Nil(t, OpenPortugueseSynthesizerFromDir(""))
	require.Nil(t, OpenPortugueseSynthesizerFromDir(t.TempDir()))
	require.Nil(t, OpenPortugueseSynthesizerFromDictPath(""))

	p := DiscoverPortugueseSynthDict()
	if p == "" {
		t.Skip("portuguese_synth.dict not discoverable")
	}
	s := OpenPortugueseSynthesizerFromDictPath(p)
	require.NotNil(t, s)
	require.NotNil(t, s.Lookup)
	require.NotEmpty(t, s.PossibleTags)
	require.Equal(t, PortugueseSynthDict, s.ResourceFileName)
	require.Equal(t, PortugueseTagsFile, s.TagFileName)
	require.Equal(t, PortugueseSorFile, s.SorFileName)
}

// Instance singleton binds INSTANCE to loaded official synthesizer.
func TestPortugueseSynthesizer_Instance(t *testing.T) {
	s := Instance()
	if s == nil {
		t.Skip("portuguese_synth.dict / official PT synth resources not available")
	}
	require.Same(t, s, Instance())
	require.Same(t, s, INSTANCE)
	require.NotNil(t, s.Lookup)
	require.NotEmpty(t, s.PossibleTags)
}
