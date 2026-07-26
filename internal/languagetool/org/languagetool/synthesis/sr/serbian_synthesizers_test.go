package sr

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSerbianSynthesizers_ResourcePaths(t *testing.T) {
	require.Equal(t, "/sr/dictionary/", DictionaryPath)
	require.Equal(t, "/sr/dictionary/serbian_synth_tags.txt", TagsFileName)
	require.Equal(t, "/sr/dictionary/ekavian/serbian_synth.dict", EkavianSynthDict)
	require.Equal(t, "/sr/dictionary/jekavian/serbian_synth.dict", JekavianSynthDict)

	require.Equal(t, EkavianSynthDict, NewEkavianSynthesizer(nil).ResourceFileName)
	require.Equal(t, TagsFileName, NewEkavianSynthesizer(nil).TagFileName)
	require.Equal(t, "sr", NewEkavianSynthesizer(nil).LangShortCode)

	require.Equal(t, JekavianSynthDict, NewJekavianSynthesizer(nil).ResourceFileName)
	require.Equal(t, TagsFileName, NewJekavianSynthesizer(nil).TagFileName)

	// Base ctor with explicit paths
	s := NewSerbianSynthesizer(EkavianSynthDict, TagsFileName, nil)
	require.Equal(t, EkavianSynthDict, s.ResourceFileName)
	require.Equal(t, TagsFileName, s.TagFileName)
	require.Equal(t, "sr", s.LangShortCode)
}
