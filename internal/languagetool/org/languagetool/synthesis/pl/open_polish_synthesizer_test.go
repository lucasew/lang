package pl

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpenPolishSynthesizerFromDir_Missing(t *testing.T) {
	require.Nil(t, OpenPolishSynthesizerFromDir(""))
	require.Nil(t, OpenPolishSynthesizerFromDir(t.TempDir()))
	require.Nil(t, OpenPolishSynthesizerFromDictPath(""))
}

func TestOpenPolishSynthesizerFromDictPath_RealDict(t *testing.T) {
	dir, _ := os.Getwd()
	var dict string
	for i := 0; i < 12; i++ {
		cand := filepath.Join(dir, "inspiration/languagetool/languagetool-language-modules/pl/src/main/resources/org/languagetool/resource/pl/polish_synth.dict")
		if st, err := os.Stat(cand); err == nil && st.Mode().IsRegular() {
			dict = cand
			break
		}
		dir = filepath.Dir(dir)
	}
	if dict == "" {
		t.Skip("polish_synth.dict not found")
	}
	s := OpenPolishSynthesizerFromDictPath(dict)
	require.NotNil(t, s)
	require.NotNil(t, s.Lookup)
	// getPosTagCorrection retained on Polish type
	require.Contains(t, s.GetPosTagCorrection("foo.bar"), ".*|.*")
}
