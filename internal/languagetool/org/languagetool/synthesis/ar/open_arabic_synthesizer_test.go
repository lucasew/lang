package ar

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpenArabicSynthesizerFromDir_Missing(t *testing.T) {
	require.Nil(t, OpenArabicSynthesizerFromDir(""))
	require.Nil(t, OpenArabicSynthesizerFromDir(t.TempDir()))
	require.Nil(t, OpenArabicSynthesizerFromDictPath(""))
}

func TestOpenArabicSynthesizerFromDictPath_RealDict(t *testing.T) {
	p := DiscoverArabicSynthDict()
	if p == "" {
		t.Skip("arabic_synth.dict not found")
	}
	s := OpenArabicSynthesizerFromDictPath(p)
	require.NotNil(t, s)
	require.NotNil(t, s.Lookup)
	require.NotEmpty(t, s.PossibleTags)
	// getPosTagCorrection must not panic / identity-only (Arabic type)
	_ = s.GetPosTagCorrection("Nxx-M1I-xW-L")
	// Java TAGS_FILE_NAME is arabic_tags.txt (not arabic_synth.dict_tags.txt)
	require.Equal(t, ArabicTagsFile, s.TagFileName)
	require.Equal(t, ArabicSynthDict, s.ResourceFileName)
}

func TestDiscoverArabicSynthDict(t *testing.T) {
	p := DiscoverArabicSynthDict()
	if p == "" {
		t.Skip("arabic_synth.dict not discoverable")
	}
	require.Contains(t, p, "arabic_synth.dict")
}
