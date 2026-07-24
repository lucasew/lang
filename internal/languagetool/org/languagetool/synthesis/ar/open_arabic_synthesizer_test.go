package ar

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpenArabicSynthesizerFromDir_Missing(t *testing.T) {
	require.Nil(t, OpenArabicSynthesizerFromDir(""))
	require.Nil(t, OpenArabicSynthesizerFromDir(t.TempDir()))
	require.Nil(t, OpenArabicSynthesizerFromDictPath(""))
}

func TestOpenArabicSynthesizerFromDictPath_RealDict(t *testing.T) {
	dir, _ := os.Getwd()
	var dict string
	for i := 0; i < 12; i++ {
		cand := filepath.Join(dir, "inspiration/languagetool/languagetool-language-modules/ar/src/main/resources/org/languagetool/resource/ar/arabic_synth.dict")
		if st, err := os.Stat(cand); err == nil && st.Mode().IsRegular() {
			dict = cand
			break
		}
		dir = filepath.Dir(dir)
	}
	if dict == "" {
		t.Skip("arabic_synth.dict not found")
	}
	s := OpenArabicSynthesizerFromDictPath(dict)
	require.NotNil(t, s)
	require.NotNil(t, s.Lookup)
	// getPosTagCorrection must not panic / identity-only (Arabic type)
	_ = s.GetPosTagCorrection("Nxx-M1I-xW-L")
}
