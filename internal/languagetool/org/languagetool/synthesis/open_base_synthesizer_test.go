package synthesis

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lucasew/lang/internal/languagetool/org/languagetool"
	"github.com/stretchr/testify/require"
)

func TestOpenBaseSynthesizer_PolishDict(t *testing.T) {
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
	s := OpenBaseSynthesizerFromDictPath("pl", dict)
	require.NotNil(t, s)
	require.NotNil(t, s.Lookup)
	// smoke: unknown tag → empty (no invent)
	pos, lemma := "subst", "x"
	tok := languagetool.NewAnalyzedToken("x", &pos, &lemma)
	forms, err := s.Synthesize(tok, "ZZZ_NO_SUCH_TAG")
	require.NoError(t, err)
	require.Empty(t, forms)
}
